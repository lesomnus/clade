package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/flg"
	"github.com/lesomnus/z"
)

func NewCmdGraph() *xli.Command {
	return &xli.Command{
		Name:  "graph",
		Brief: "print the dependency graph as a tree",

		Flags: flg.Flags{
			&flg.String{Name: "ports", Brief: "path to the ports directory"},
			&flg.String{Name: "graph", Brief: "read a serialized graph (.json or binary) instead of recomputing"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			flg.VisitP(cmd, "ports", &c.Ports)

			g, err := obtainGraph(ctx, c, cmd)
			if err != nil {
				return z.Err(err, "obtain graph")
			}

			return renderTree(cmd, g.Nodes)
		}),
	}
}

// renderTree prints the graph as an indented tree. Roots are the external
// upstream images (and any baseless nodes, e.g. http sources); each node is
// nested under the base it derives from. Outdated targets are flagged.
func renderTree(w io.Writer, nodes []*cladev1.Node) error {
	bw := bufio.NewWriter(w)

	by_id := map[string]*cladev1.Node{}
	for _, n := range nodes {
		by_id[n.Id] = n
	}

	children := map[string][]*cladev1.Node{}
	for _, n := range nodes {
		children[n.Base] = append(children[n.Base], n)
	}

	faint := color.New(color.Faint).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	label := func(ref string) string {
		n, ok := by_id[ref]
		if !ok {
			return faint(ref + " (external)")
		}
		s := n.Id
		if len(n.Tags) > 1 {
			extra := make([]string, 0, len(n.Tags)-1)
			for _, t := range n.Tags[1:] {
				extra = append(extra, "+"+tagOf(t))
			}
			s += " " + faint(strings.Join(extra, " "))
		}
		if n.Outdated {
			s += " " + red("[outdated]")
		} else {
			s += " " + green("[ok]")
		}
		return s
	}

	var walk func(ref, prefix string, last bool)
	walk = func(ref, prefix string, last bool) {
		connector, next := "├─ ", prefix+"│  "
		if last {
			connector, next = "└─ ", prefix+"   "
		}
		fmt.Fprintf(bw, "%s%s%s\n", prefix, connector, label(ref))

		kids := children[ref]
		for i, k := range kids {
			walk(k.Id, next, i == len(kids)-1)
		}
	}

	// Roots have children but are not children themselves: external upstreams
	// plus any node with no base at all (e.g. an http source).
	roots := externalBases(nodes)
	for _, n := range nodes {
		if n.Base == "" {
			roots = append(roots, n.Id)
		}
	}

	for _, r := range roots {
		fmt.Fprintln(bw, label(r))
		kids := children[r]
		for i, k := range kids {
			walk(k.Id, "", i == len(kids)-1)
		}
	}

	return bw.Flush()
}

// externalBases lists, in graph order, the base references that are not produced
// by any node (i.e. upstream images clade only consumes). Each appears once.
func externalBases(nodes []*cladev1.Node) []string {
	is_node := map[string]bool{}
	for _, n := range nodes {
		is_node[n.Id] = true
	}

	seen := map[string]bool{}
	out := []string{}
	for _, n := range nodes {
		if n.Base != "" && !is_node[n.Base] && !seen[n.Base] {
			seen[n.Base] = true
			out = append(out, n.Base)
		}
	}
	return out
}

// tagOf returns the tag portion of a "repo:tag" reference.
func tagOf(ref string) string {
	if i := strings.LastIndex(ref, ":"); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

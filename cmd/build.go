package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lesomnus/clade/builder"
	"github.com/lesomnus/clade/cmd/config"
	"github.com/lesomnus/clade/compare"
	"github.com/lesomnus/clade/graph"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
	"github.com/lesomnus/clade/registry"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/arg"
	"github.com/lesomnus/xli/flg"
	"github.com/lesomnus/z"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// baseNameLabel records the upstream reference a target was built from.
const baseNameLabel = "org.opencontainers.image.base.name"

func NewCmdBuild() *xli.Command {
	return &xli.Command{
		Name:  "build",
		Brief: "build outdated images following the dependency graph",

		Args: arg.Args{
			&arg.RestStrings{Name: "node", Brief: "node ids to build (default: all outdated)"},
		},
		Flags: flg.Flags{
			&flg.String{Name: "ports", Brief: "path to the ports directory"},
			&flg.String{Name: "graph", Brief: "read a serialized graph (.json or binary) instead of recomputing"},
			&flg.Switch{Name: "all", Brief: "build every node in the graph, not only outdated ones"},
			&flg.Switch{Name: "no-push", Brief: "do not push built images"},
			&flg.Switch{Name: "load", Brief: "load built images into the local docker store (implies no push)"},
			&flg.Switch{Name: "dry-run", Brief: "print build commands without executing them"},
			&flg.String{Name: "docker", Brief: "docker binary to invoke (default docker)"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			flg.VisitP(cmd, "ports", &c.Ports)
			flg.VisitP(cmd, "docker", &c.Build.Docker)

			g, err := obtainGraph(ctx, c, cmd)
			if err != nil {
				return z.Err(err, "obtain graph")
			}

			all := false
			flg.VisitP(cmd, "all", &all)
			ids, _ := arg.Get[[]string](cmd, "node")

			targets, err := selectBuildTargets(g, ids, all)
			if err != nil {
				return z.Err(err, "select targets")
			}
			if len(targets) == 0 {
				cmd.Println("nothing to build")
				return nil
			}

			no_push := false
			flg.VisitP(cmd, "no-push", &no_push)
			load := false
			flg.VisitP(cmd, "load", &load)
			dry_run := false
			flg.VisitP(cmd, "dry-run", &dry_run)

			runner := &buildRunner{
				reg:        registry.NewRemote(), // fresh: a just-pushed base must resolve
				loadPort:   port.Load,
				newBuilder: builder.New,
				push:       !no_push && !load,
				load:       load,
				dryRun:     dry_run,
				bin:        c.Build.Docker,
				stdout:     cmd,
				stderr:     os.Stderr,
			}
			return runner.run(ctx, targets)
		}),
	}
}

// obtainGraph reads a serialized graph when --graph is set, otherwise recomputes
// it from the ports (the same pipeline as `outdated`).
func obtainGraph(ctx context.Context, c *config.Config, cmd *xli.Command) (*cladev1.Graph, error) {
	if p, ok := flg.Get[string](cmd, "graph"); ok && p != "" {
		return readGraphFile(p)
	}

	reg, err := buildRegistry(c)
	if err != nil {
		return nil, z.Err(err, "build registry")
	}
	ports, err := port.LoadAll(c.Ports)
	if err != nil {
		return nil, z.Err(err, "load ports")
	}

	b := &graph.Builder{Registry: reg}
	return b.Build(ctx, ports)
}

func readGraphFile(path string) (*cladev1.Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read graph %q: %w", path, err)
	}

	g := &cladev1.Graph{}
	if strings.HasSuffix(path, ".json") {
		if err := protojson.Unmarshal(data, g); err != nil {
			return nil, fmt.Errorf("decode json graph: %w", err)
		}
	} else if err := proto.Unmarshal(data, g); err != nil {
		return nil, fmt.Errorf("decode binary graph: %w", err)
	}
	return g, nil
}

// selectBuildTargets resolves which nodes to build, preserving the graph's
// topological order. With explicit ids, exactly those are built; otherwise the
// outdated nodes (or all nodes when all is true).
func selectBuildTargets(g *cladev1.Graph, ids []string, all bool) ([]*cladev1.Node, error) {
	if len(ids) > 0 {
		want := map[string]bool{}
		for _, id := range ids {
			want[id] = true
		}

		out := make([]*cladev1.Node, 0, len(ids))
		for _, n := range g.Nodes {
			if want[n.Id] {
				out = append(out, n)
				delete(want, n.Id)
			}
		}
		if len(want) > 0 {
			missing := make([]string, 0, len(want))
			for id := range want {
				missing = append(missing, id)
			}
			return nil, fmt.Errorf("unknown node(s): %s", strings.Join(missing, ", "))
		}
		return out, nil
	}

	out := make([]*cladev1.Node, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if all || n.Outdated {
			out = append(out, n)
		}
	}
	return out, nil
}

// buildRunner builds a topologically ordered list of nodes, constructing a
// builder per node from its port's build config.
type buildRunner struct {
	reg        registry.Registry
	loadPort   func(dir string) (*port.Port, error)
	newBuilder func(kind string, params []byte, spec builder.Spec) (builder.Builder, error)

	push   bool
	load   bool
	dryRun bool
	bin    string
	stdout io.Writer
	stderr io.Writer
}

func (r *buildRunner) run(ctx context.Context, targets []*cladev1.Node) error {
	ports := map[string]*port.Port{}
	for _, node := range targets {
		p, ok := ports[node.Port]
		if !ok {
			var err error
			p, err = r.loadPort(node.Port)
			if err != nil {
				return z.Err(err, "load port %q", node.Port)
			}
			ports[node.Port] = p
		}

		bld, err := r.newBuilder(p.Build.Kind, p.Build.Params, r.spec(ctx, node))
		if err != nil {
			return z.Err(err, "builder for %q", node.Id)
		}
		if err := bld.Build(ctx); err != nil {
			return z.Err(err, "build %q", node.Id)
		}
	}
	return nil
}

// spec builds the runtime build description for a node. The upstream name and
// digest are recorded as labels so the digest comparator can detect future
// upstream changes; the digest is resolved fresh so a just-pushed base counts.
// A node without a base (e.g. an http source) records no base labels.
func (r *buildRunner) spec(ctx context.Context, node *cladev1.Node) builder.Spec {
	labels := map[string]string{}
	if node.Base != "" {
		labels[baseNameLabel] = node.Base
		if info, err := r.reg.Stat(ctx, node.Base); err == nil {
			labels[compare.DefaultBaseDigestLabel] = info.Digest
		}
	}

	return builder.Spec{
		Dir:     node.Port,
		Tags:    node.Tags,
		Base:    node.Base,
		BaseTag: node.BaseTag,
		Labels:  labels,
		Push:    r.push,
		Load:    r.load,
		DryRun:  r.dryRun,
		Bin:     r.bin,
		Stdout:  r.stdout,
		Stderr:  r.stderr,
	}
}

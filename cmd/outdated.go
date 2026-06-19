package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lesomnus/clade/cmd/config"
	"github.com/lesomnus/clade/compare"
	"github.com/lesomnus/clade/graph"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
	"github.com/lesomnus/clade/registry"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/flg"
	"github.com/lesomnus/z"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func NewCmdOutdated() *xli.Command {
	return &xli.Command{
		Name:  "outdated",
		Brief: "list build targets that are out of date with their upstream",

		Flags: flg.Flags{
			&flg.String{Name: "ports", Brief: "path to the ports directory"},
			&flg.String{Name: "compare", Brief: "outdated comparison strategy (created, digest)"},
			&flg.String{Name: "format", Brief: "output format: text, json, binary"},
			&flg.Switch{Name: "all", Brief: "include up-to-date targets"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			flg.VisitP(cmd, "ports", &c.Ports)
			flg.VisitP(cmd, "compare", &c.Compare.Kind)

			reg, err := buildRegistry(c)
			if err != nil {
				return z.Err(err, "build registry")
			}

			cmp, err := compare.New(c.Compare.Kind, c.Compare.Params)
			if err != nil {
				return z.Err(err, "build comparator")
			}

			ports, err := port.LoadAll(c.Ports)
			if err != nil {
				return z.Err(err, "load ports")
			}

			b := &graph.Builder{Registry: reg, Comparator: cmp}
			g, err := b.Build(ctx, ports)
			if err != nil {
				return z.Err(err, "build graph")
			}

			all := false
			flg.VisitP(cmd, "all", &all)

			format := "text"
			flg.VisitP(cmd, "format", &format)

			return renderGraph(cmd, selectNodes(g, all), format)
		}),
	}
}

// buildRegistry constructs a remote registry wrapped with a metadata cache.
func buildRegistry(c *config.Config) (registry.Registry, error) {
	ttl, err := time.ParseDuration(c.Cache.TTL)
	if err != nil {
		return nil, fmt.Errorf("parse cache ttl %q: %w", c.Cache.TTL, err)
	}

	dir := c.Cache.Dir
	if dir == "" {
		if base, err := os.UserCacheDir(); err == nil {
			dir = filepath.Join(base, "clade")
		}
	}

	var cache registry.Cache
	if dir == "" {
		cache = registry.NewMemCache()
	} else {
		fc, err := registry.NewFileCache(dir)
		if err != nil {
			return nil, err
		}
		cache = fc
	}

	return registry.WithCache(registry.NewRemote(), cache, ttl), nil
}

// selectNodes returns the outdated nodes, or all nodes when all is true.
func selectNodes(g *cladev1.Graph, all bool) []*cladev1.Node {
	if all {
		return g.Nodes
	}

	out := make([]*cladev1.Node, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.Outdated {
			out = append(out, n)
		}
	}
	return out
}

func renderGraph(cmd *xli.Command, nodes []*cladev1.Node, format string) error {
	g := &cladev1.Graph{Nodes: nodes}

	switch format {
	case "", "text":
		for _, n := range nodes {
			status := "ok"
			if n.Outdated {
				status = "outdated"
			}
			cmd.Printf("%-9s %s  (base: %s)\n", status, n.Id, n.Base)
		}
		return nil

	case "json":
		b, err := protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(g)
		if err != nil {
			return z.Err(err, "marshal json")
		}
		cmd.Println(string(b))
		return nil

	case "binary":
		b, err := proto.Marshal(g)
		if err != nil {
			return z.Err(err, "marshal binary")
		}
		_, err = cmd.Write(b)
		return err

	default:
		return fmt.Errorf("unknown format %q", format)
	}
}

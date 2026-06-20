package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/lesomnus/clade/cmd/config"
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
			&flg.String{Name: "format", Brief: "output format: text, json, binary"},
			&flg.Switch{Name: "all", Brief: "include up-to-date targets"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			flg.VisitP(cmd, "ports", &c.Ports)

			reg, err := buildRegistry(c)
			if err != nil {
				return z.Err(err, "build registry")
			}

			ports, err := port.LoadAll(c.Ports)
			if err != nil {
				return z.Err(err, "load ports")
			}

			b := &graph.Builder{Registry: reg}
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

	dir := cacheDir(c)

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

// cacheDir resolves the on-disk cache directory from config, falling back to
// <user cache dir>/clade. An empty result means no suitable directory was found
// (only possible when os.UserCacheDir fails and cache.dir is unset).
func cacheDir(c *config.Config) string {
	if c.Cache.Dir != "" {
		return c.Cache.Dir
	}
	if base, err := os.UserCacheDir(); err == nil {
		return filepath.Join(base, "clade")
	}
	return ""
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
			tags := n.Tags
			if len(tags) == 0 {
				tags = []string{n.Id}
			}
			cmd.Printf("%s from %s\n", status, color.New(color.Underline).Sprint(n.Base))
			for _, t := range tags {
				cmd.Printf("\t%s\n", color.New(color.Bold).Sprint(t))
			}
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

package graph_test

import (
	"context"
	"testing"
	"time"

	"github.com/lesomnus/clade/compare"
	"github.com/lesomnus/clade/graph"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
	"github.com/lesomnus/clade/registry"
)

func semverPort(dir, parentRepo, buildRepo string) *port.Port {
	return &port.Port{
		Dir: dir,
		Parent: port.Parent{
			Repo:   parentRepo,
			Target: port.Target{Kind: "semver", Params: []byte("kind: semver\n")},
		},
		Build: port.Build{Repo: buildRepo, Tag: "{{.Major}}.{{.Minor}}.{{.Patch}}"},
	}
}

func at(sec int64) time.Time { return time.Unix(sec, 0) }

func nodeByID(g *cladev1.Graph, id string) *cladev1.Node {
	for _, n := range g.Nodes {
		if n.Id == id {
			return n
		}
	}
	return nil
}

func TestBuildGraph(t *testing.T) {
	reg := registry.NewFake()
	// External upstream.
	reg.Set("up.io/base:1.0.0", &registry.ImageInfo{Created: at(100)})
	reg.Set("up.io/base:1.1.0", &registry.ImageInfo{Created: at(100)})
	reg.Set("up.io/base:2.0.0", &registry.ImageInfo{Created: at(100)})
	// Port A targets.
	reg.Set("me.io/a:1.0.0", &registry.ImageInfo{Created: at(200)}) // newer than base -> ok
	reg.Set("me.io/a:1.1.0", &registry.ImageInfo{Created: at(50)})  // older than base -> outdated
	// me.io/a:2.0.0 is missing -> outdated.
	// Port B targets (parent is me.io/a, an internal edge).
	reg.Set("me.io/b:1.0.0", &registry.ImageInfo{Created: at(300)}) // parent ok, newer -> ok
	reg.Set("me.io/b:1.1.0", &registry.ImageInfo{Created: at(300)}) // parent outdated -> propagated
	// me.io/b:2.0.0 is missing.

	ports := []*port.Port{
		semverPort("ports/a", "up.io/base", "me.io/a"),
		semverPort("ports/b", "me.io/a", "me.io/b"),
	}

	cmp, err := compare.New("created", nil)
	if err != nil {
		t.Fatal(err)
	}
	b := &graph.Builder{Registry: reg, Comparator: cmp}
	g, err := b.Build(context.Background(), ports)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	if len(g.Nodes) != 6 {
		t.Fatalf("nodes = %d, want 6", len(g.Nodes))
	}

	want := map[string]bool{
		"me.io/a:1.0.0": false,
		"me.io/a:1.1.0": true,
		"me.io/a:2.0.0": true,
		"me.io/b:1.0.0": false,
		"me.io/b:1.1.0": true, // propagated from parent
		"me.io/b:2.0.0": true, // propagated from parent
	}
	for id, outdated := range want {
		n := nodeByID(g, id)
		if n == nil {
			t.Fatalf("missing node %q", id)
		}
		if n.Outdated != outdated {
			t.Errorf("%s outdated = %v, want %v", id, n.Outdated, outdated)
		}
	}

	// Internal edge: a b-node points at the matching a-node.
	b11 := nodeByID(g, "me.io/b:1.1.0")
	if b11.Base != "me.io/a:1.1.0" {
		t.Errorf("b:1.1.0 base = %q, want me.io/a:1.1.0", b11.Base)
	}
	if len(b11.Parents) != 1 || b11.Parents[0] != "me.io/a:1.1.0" {
		t.Errorf("b:1.1.0 parents = %v, want [me.io/a:1.1.0]", b11.Parents)
	}

	// External base: no internal parent.
	a10 := nodeByID(g, "me.io/a:1.0.0")
	if len(a10.Parents) != 0 {
		t.Errorf("a:1.0.0 parents = %v, want none", a10.Parents)
	}
	if a10.Base != "up.io/base:1.0.0" {
		t.Errorf("a:1.0.0 base = %q", a10.Base)
	}

	// Topological order: every parent precedes its children.
	pos := map[string]int{}
	for i, n := range g.Nodes {
		pos[n.Id] = i
	}
	for _, n := range g.Nodes {
		for _, p := range n.Parents {
			if pos[p] >= pos[n.Id] {
				t.Errorf("parent %q is not ordered before child %q", p, n.Id)
			}
		}
	}
}

func TestBuildCycle(t *testing.T) {
	reg := registry.NewFake()
	ports := []*port.Port{
		semverPort("ports/a", "me.io/b", "me.io/a"),
		semverPort("ports/b", "me.io/a", "me.io/b"),
	}
	cmp, _ := compare.New("created", nil)
	b := &graph.Builder{Registry: reg, Comparator: cmp}
	if _, err := b.Build(context.Background(), ports); err == nil {
		t.Fatal("expected a cycle error")
	}
}

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesomnus/clade/builder"
	"github.com/lesomnus/clade/compare"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
	"github.com/lesomnus/clade/registry"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func node(id, base, portDir string, outdated bool, parents ...string) *cladev1.Node {
	return &cladev1.Node{Id: id, Tags: []string{id}, Base: base, Port: portDir, Outdated: outdated, Parents: parents}
}

func ids(nodes []*cladev1.Node) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.Id
	}
	return out
}

func eq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sampleGraph() *cladev1.Graph {
	return &cladev1.Graph{Nodes: []*cladev1.Node{
		node("a:1", "up:1", "ports/a", false),
		node("b:1", "a:1", "ports/b", true, "a:1"),
		node("b:2", "a:2", "ports/b", true),
	}}
}

func TestSelectBuildTargets(t *testing.T) {
	g := sampleGraph()

	if got := mustSelect(t, g, nil, false); !eq(ids(got), []string{"b:1", "b:2"}) {
		t.Errorf("default = %v, want [b:1 b:2]", ids(got))
	}
	if got := mustSelect(t, g, nil, true); !eq(ids(got), []string{"a:1", "b:1", "b:2"}) {
		t.Errorf("all = %v, want [a:1 b:1 b:2]", ids(got))
	}
	// Order follows the graph, not the argument order.
	if got := mustSelect(t, g, []string{"b:2", "a:1"}, false); !eq(ids(got), []string{"a:1", "b:2"}) {
		t.Errorf("ids = %v, want [a:1 b:2]", ids(got))
	}
	if _, err := selectBuildTargets(g, []string{"nope"}, false); err == nil {
		t.Error("expected error for unknown node")
	}
}

func mustSelect(t *testing.T, g *cladev1.Graph, ids []string, all bool) []*cladev1.Node {
	t.Helper()
	out, err := selectBuildTargets(g, ids, all)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestBuildRunner(t *testing.T) {
	reg := registry.NewFake()
	reg.Set("a:1", &registry.ImageInfo{Digest: "sha256:base1"})

	ports := map[string]*port.Port{
		"ports/b": {
			Dir: "ports/b",
			Build: port.Build{
				Repo:   "b",
				Tags:   []string{"{{.Major}}"},
				Kind:   "build",
				Params: []byte("platforms: [linux/amd64]\n"),
			},
		},
	}

	var fakes []*builder.Fake
	loads := 0
	runner := &buildRunner{
		reg: reg,
		loadPort: func(dir string) (*port.Port, error) {
			loads++
			return ports[dir], nil
		},
		newBuilder: builder.NewFake(&fakes),
		push:       true,
	}

	targets := []*cladev1.Node{
		node("b:1", "a:1", "ports/b", true),
		node("b:2", "a:2", "ports/b", true), // shares the port -> loaded once
	}
	if err := runner.run(context.Background(), targets); err != nil {
		t.Fatal(err)
	}

	if loads != 1 {
		t.Errorf("loadPort called %d times, want 1 (cached)", loads)
	}
	if len(fakes) != 2 {
		t.Fatalf("got %d builds, want 2", len(fakes))
	}

	s0 := fakes[0].Spec
	if !eq(s0.Tags, []string{"b:1"}) {
		t.Errorf("tags = %v, want [b:1]", s0.Tags)
	}
	if s0.Dir != "ports/b" {
		t.Errorf("dir = %q", s0.Dir)
	}
	if s0.Base != "a:1" || !s0.Push {
		t.Errorf("base=%q push=%v", s0.Base, s0.Push)
	}
	if s0.Labels[baseNameLabel] != "a:1" {
		t.Errorf("base name label = %q", s0.Labels[baseNameLabel])
	}
	if s0.Labels[compare.DefaultBaseDigestLabel] != "sha256:base1" {
		t.Errorf("base digest label = %q", s0.Labels[compare.DefaultBaseDigestLabel])
	}
	// The port's raw build options reach the builder via params.
	if !strings.Contains(string(fakes[0].Params), "platforms") {
		t.Errorf("params not passed through: %q", fakes[0].Params)
	}
	if fakes[0].Built != 1 {
		t.Errorf("Build called %d times, want 1", fakes[0].Built)
	}

	// Second target's base (a:2) is absent from the registry, so no digest label.
	if _, ok := fakes[1].Spec.Labels[compare.DefaultBaseDigestLabel]; ok {
		t.Errorf("unexpected base digest label for absent base: %v", fakes[1].Spec.Labels)
	}
}

func TestReadGraphFile(t *testing.T) {
	g := sampleGraph()
	dir := t.TempDir()

	bin, err := proto.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(dir, "g.pb")
	if err := os.WriteFile(binPath, bin, 0o644); err != nil {
		t.Fatal(err)
	}

	js, err := protojson.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(dir, "g.json")
	if err := os.WriteFile(jsonPath, js, 0o644); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{binPath, jsonPath} {
		got, err := readGraphFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if len(got.Nodes) != 3 {
			t.Errorf("%s: nodes = %d, want 3", path, len(got.Nodes))
		}
	}
}

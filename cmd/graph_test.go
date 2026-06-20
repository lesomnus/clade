package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fatih/color"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
)

func TestRenderTree(t *testing.T) {
	color.NoColor = true // deterministic, no ANSI codes

	// up:1 -> a:1 -> b:1 (outdated), and a separate external a:2 -> b:2.
	g := &cladev1.Graph{Nodes: []*cladev1.Node{
		{Id: "a:1", Tags: []string{"a:1", "a:1.0"}, Base: "up:1"},
		{Id: "b:1", Tags: []string{"b:1"}, Base: "a:1", Outdated: true, Parents: []string{"a:1"}},
		{Id: "b:2", Tags: []string{"b:2"}, Base: "a:2", Outdated: true},
	}}

	var buf bytes.Buffer
	if err := renderTree(&buf, g.Nodes); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	want := strings.Join([]string{
		"up:1 (external)",
		"└─ a:1 +1.0 [ok]",
		"   └─ b:1 [outdated]",
		"a:2 (external)",
		"└─ b:2 [outdated]",
		"",
	}, "\n")

	if out != want {
		t.Errorf("tree mismatch\n got:\n%s\nwant:\n%s", out, want)
	}
}

func TestRenderTreeBranching(t *testing.T) {
	color.NoColor = true

	// One external base with two children; the second child has its own child.
	g := &cladev1.Graph{Nodes: []*cladev1.Node{
		{Id: "a:1", Tags: []string{"a:1"}, Base: "up:1"},
		{Id: "a:2", Tags: []string{"a:2"}, Base: "up:1"},
		{Id: "b:1", Tags: []string{"b:1"}, Base: "a:2", Outdated: true, Parents: []string{"a:2"}},
	}}

	var buf bytes.Buffer
	if err := renderTree(&buf, g.Nodes); err != nil {
		t.Fatal(err)
	}

	want := strings.Join([]string{
		"up:1 (external)",
		"├─ a:1 [ok]",
		"└─ a:2 [ok]",
		"   └─ b:1 [outdated]",
		"",
	}, "\n")

	if got := buf.String(); got != want {
		t.Errorf("tree mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderTreeBaselessRoot(t *testing.T) {
	color.NoColor = true

	// An http source has no base, so its node is a root itself.
	g := &cladev1.Graph{Nodes: []*cladev1.Node{
		{Id: "ghcr.io/me/claude:1.2.3", Tags: []string{"ghcr.io/me/claude:1.2.3"}, Base: "", Outdated: true},
	}}

	var buf bytes.Buffer
	if err := renderTree(&buf, g.Nodes); err != nil {
		t.Fatal(err)
	}

	want := "ghcr.io/me/claude:1.2.3 [outdated]\n"
	if got := buf.String(); got != want {
		t.Errorf("tree mismatch\n got:\n%s\nwant:\n%s", got, want)
	}
}

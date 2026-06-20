package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fatih/color"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
)

func TestHyperlink(t *testing.T) {
	got := hyperlink("file:///x/port.yaml", "name")
	want := "\x1b]8;;file:///x/port.yaml\x1b\\name\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("hyperlink = %q, want %q", got, want)
	}
}

func TestPortLabel(t *testing.T) {
	defer func(prev bool) { color.NoColor = prev }(color.NoColor)
	color.NoColor = true // strip styling for deterministic assertions

	ports := map[string]*port.Port{
		"ports/dev-golang": {Dir: "ports/dev-golang", Name: "dev-golang"},
		"ports/claude":     {Dir: "ports/claude", Name: "my-claude"},
	}

	if got := portLabel("ports/claude", ports, false); got != "my-claude" {
		t.Errorf("label = %q, want my-claude", got)
	}
	// Unknown port falls back to the directory base name.
	if got := portLabel("ports/unknown", ports, false); got != "unknown" {
		t.Errorf("label = %q, want unknown", got)
	}
	// With link, the name is wrapped in an OSC 8 file:// hyperlink.
	got := portLabel("ports/dev-golang", ports, true)
	abs, _ := filepath.Abs(filepath.Join("ports/dev-golang", port.Filename))
	want := hyperlink("file://"+abs, "dev-golang")
	if got != want {
		t.Errorf("linked label = %q, want %q", got, want)
	}
}

func TestRenderText(t *testing.T) {
	defer func(prev bool) { color.NoColor = prev }(color.NoColor)
	color.NoColor = true

	ports := map[string]*port.Port{
		"ports/dev-golang": {Dir: "ports/dev-golang", Name: "dev-golang"},
		"ports/claude":     {Dir: "ports/claude", Name: "claude"},
	}
	nodes := []*cladev1.Node{
		{Id: "ghcr.io/me/dev-golang:1.24.0", Tags: []string{"ghcr.io/me/dev-golang:1.24.0"}, Base: "docker.io/library/golang:1.24", Port: "ports/dev-golang", Outdated: true},
		{Id: "ghcr.io/me/claude:1.2.3", Tags: []string{"ghcr.io/me/claude:1.2.3"}, Base: "", Port: "ports/claude", Outdated: true},
	}

	var buf bytes.Buffer
	renderText(&buf, nodes, ports, false)
	out := buf.String()

	if !strings.Contains(out, "outdated  dev-golang ports/dev-golang/port.yaml from docker.io/library/golang:1.24\n") {
		t.Errorf("container line missing name/path/base: %q", out)
	}
	// An http node has no base, so the "from <base>" part is omitted.
	if !strings.Contains(out, "outdated  claude ports/claude/port.yaml\n") {
		t.Errorf("http line should show name and path without base: %q", out)
	}
	if !strings.Contains(out, "\tghcr.io/me/claude:1.2.3\n") {
		t.Errorf("tags should be listed indented: %q", out)
	}
}

package port_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesomnus/clade/port"
)

const sample = `parent:
  repo: docker.io/library/golang
  target:
    kind: semver
    last-major: 2
    last-minor: 3
    match: "-alpine$"
build:
  repo: my-registry/golang-dev
  tag: "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
`

func writePort(t *testing.T, dir, manifest string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, port.Filename), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "golang-dev")
	writePort(t, dir, sample)

	p, err := port.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if p.Dir != dir {
		t.Errorf("dir = %q, want %q", p.Dir, dir)
	}
	if p.Parent.Repo != "docker.io/library/golang" {
		t.Errorf("parent.repo = %q", p.Parent.Repo)
	}
	if p.Parent.Target.Kind != "semver" {
		t.Errorf("target.kind = %q", p.Parent.Target.Kind)
	}
	if !strings.Contains(string(p.Parent.Target.Params), "last-major") {
		t.Errorf("target.params missing kind-specific fields: %q", p.Parent.Target.Params)
	}
	if p.Build.Repo != "my-registry/golang-dev" {
		t.Errorf("build.repo = %q", p.Build.Repo)
	}
	if p.Build.Tag != "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine" {
		t.Errorf("build.tag = %q", p.Build.Tag)
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "broken")
	writePort(t, dir, "build:\n  repo: x\n  tag: y\n") // no parent
	if _, err := port.Load(dir); err == nil {
		t.Fatal("expected error for missing parent")
	}
}

func TestLoadAll(t *testing.T) {
	root := t.TempDir()
	writePort(t, filepath.Join(root, "a"), sample)
	writePort(t, filepath.Join(root, "b"), sample)
	// A directory without a manifest is ignored.
	if err := os.MkdirAll(filepath.Join(root, "not-a-port"), 0o755); err != nil {
		t.Fatal(err)
	}

	ports, err := port.LoadAll(root)
	if err != nil {
		t.Fatalf("load all: %v", err)
	}
	if len(ports) != 2 {
		t.Fatalf("loaded %d ports, want 2", len(ports))
	}
	if ports[0].Dir >= ports[1].Dir {
		t.Errorf("ports not sorted by dir: %q, %q", ports[0].Dir, ports[1].Dir)
	}
}

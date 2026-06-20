package port_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesomnus/clade/port"
)

const sample = `source:
  kind: container
  repo: docker.io/library/golang
select:
  kind: semver
  last-major: 2
  last-minor: 3
  pre-release: alpine
compare:
  - kind: created
  - kind: digest
build:
  repo: my-registry/dev-golang
  tags:
    - "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
    - "{{.Major}}.{{.Minor}}-alpine"
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
	dir := filepath.Join(t.TempDir(), "dev-golang")
	writePort(t, dir, sample)

	p, err := port.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if p.Dir != dir {
		t.Errorf("dir = %q, want %q", p.Dir, dir)
	}
	if p.Source.Kind != "container" {
		t.Errorf("source.kind = %q", p.Source.Kind)
	}
	if p.Source.Repo != "docker.io/library/golang" {
		t.Errorf("source.repo = %q", p.Source.Repo)
	}
	if p.Select.Kind != "semver" {
		t.Errorf("select.kind = %q", p.Select.Kind)
	}
	if !strings.Contains(string(p.Select.Params), "last-major") {
		t.Errorf("select.params missing kind-specific fields: %q", p.Select.Params)
	}
	if len(p.Compare) != 2 || p.Compare[0].Kind != "created" || p.Compare[1].Kind != "digest" {
		t.Errorf("compare = %+v, want [created digest]", p.Compare)
	}
	if p.Build.Repo != "my-registry/dev-golang" {
		t.Errorf("build.repo = %q", p.Build.Repo)
	}
	if len(p.Build.Tags) != 2 || p.Build.Tags[0] != "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine" || p.Build.Tags[1] != "{{.Major}}.{{.Minor}}-alpine" {
		t.Errorf("build.tags = %q", p.Build.Tags)
	}
}

func TestLoadHTTPSource(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "claude")
	writePort(t, dir, `source:
  kind: http
  url: https://example.com/stable
select:
  kind: semver
build:
  repo: my-registry/claude
  tags:
    - "{{.Major}}.{{.Minor}}.{{.Patch}}"
`)

	p, err := port.Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if p.Source.Kind != "http" || p.Source.Url != "https://example.com/stable" {
		t.Errorf("source = %+v", p.Source)
	}
	// compare is optional and absent here.
	if p.Compare != nil {
		t.Errorf("compare = %v, want nil", p.Compare)
	}
}

func TestLoadInvalid(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "broken")
	writePort(t, dir, "build:\n  repo: x\n  tags: [y]\n") // no source
	if _, err := port.Load(dir); err == nil {
		t.Fatal("expected error for missing source")
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

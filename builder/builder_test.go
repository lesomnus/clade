package builder_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/lesomnus/clade/builder"
)

// buildAndCapture constructs the builder of kind, runs a dry-run build, and
// returns what it wrote.
func buildAndCapture(t *testing.T, kind, params string, spec builder.Spec) string {
	t.Helper()
	var buf bytes.Buffer
	spec.DryRun = true
	spec.Stdout = &buf
	spec.Stderr = &buf

	b, err := builder.New(kind, []byte(params), spec)
	if err != nil {
		t.Fatalf("new %q: %v", kind, err)
	}
	if err := b.Build(context.Background()); err != nil {
		t.Fatalf("build: %v", err)
	}
	return buf.String()
}

const sampleParams = `
dockerfile: Dockerfile
target: final
platforms: [linux/amd64, linux/arm64]
args:
  FOO: bar
labels:
  a: b
cache-from: [type=gha]
no-cache: true
pull: true
extra-args: ["--quiet"]
`

func sampleSpec() builder.Spec {
	return builder.Spec{
		Dir:    "ports/x",
		Tags:   []string{"repo:1", "repo:latest"},
		Base:   "up:1",
		Labels: map[string]string{"base": "x"},
		Push:   true,
	}
}

func TestBuildxArgv(t *testing.T) {
	out := buildAndCapture(t, "build", sampleParams, sampleSpec())

	if !strings.HasPrefix(out, "docker buildx build ") {
		t.Fatalf("unexpected prefix: %s", out)
	}
	for _, want := range []string{
		"--file ports/x/Dockerfile",
		"--target final",
		"--platform linux/amd64",
		"--platform linux/arm64",
		"--build-arg BASE=up:1", // injected from spec.Base
		"--build-arg FOO=bar",
		"--label a=b",
		"--label base=x", // injected from spec.Labels
		"--cache-from type=gha",
		"--no-cache",
		"--pull",
		"--tag repo:1",
		"--tag repo:latest",
		"--push",
		"--quiet",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("argv missing %q in: %s", want, out)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(out), "ports/x") {
		t.Errorf("context should be the last arg: %s", out)
	}
	// injected build-arg keys are sorted (BASE before FOO).
	if strings.Index(out, "BASE=up:1") > strings.Index(out, "FOO=bar") {
		t.Errorf("build-arg not sorted: %s", out)
	}
}

func TestKindDefaultsToBuild(t *testing.T) {
	out := buildAndCapture(t, "", "", builder.Spec{Dir: ".", Tags: []string{"x:1"}, Base: "b:1"})
	if !strings.Contains(out, "docker buildx build") {
		t.Errorf("empty kind should default to build: %s", out)
	}
}

func TestUnknownKind(t *testing.T) {
	if _, err := builder.New("nope", nil, builder.Spec{}); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestPushLoadConflict(t *testing.T) {
	for _, kind := range []string{"build", "bake"} {
		b, err := builder.New(kind, nil, builder.Spec{Push: true, Load: true})
		if err != nil {
			t.Fatal(err)
		}
		if err := b.Build(context.Background()); err == nil {
			t.Errorf("%s: expected push+load error", kind)
		}
	}
}

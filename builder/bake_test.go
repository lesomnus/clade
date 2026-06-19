package builder_test

import (
	"strings"
	"testing"

	"github.com/lesomnus/clade/builder"
)

func TestBakeDefinition(t *testing.T) {
	out := buildAndCapture(t, "bake", sampleParams, sampleSpec())

	for _, want := range []string{
		`"target"`,
		`"default"`,
		`"context": "ports/x"`,
		`"dockerfile": "ports/x/Dockerfile"`,
		`"repo:1"`,
		`"repo:latest"`,
		`"BASE": "up:1"`, // injected from spec.Base
		`"FOO": "bar"`,
		`"base": "x"`, // injected label from spec
		`"a": "b"`,
		`"cache-from"`,
		`"no-cache": true`,
		`"type=registry"`, // push output
	} {
		if !strings.Contains(out, want) {
			t.Errorf("bake definition missing %q in: %s", want, out)
		}
	}

	if !strings.Contains(out, "docker buildx bake --file") {
		t.Errorf("missing bake command: %s", out)
	}
	if !strings.Contains(out, "--quiet") { // extra-args
		t.Errorf("missing extra-args: %s", out)
	}
}

func TestBakeLoadOutput(t *testing.T) {
	spec := builder.Spec{Dir: ".", Tags: []string{"x:1"}, Base: "b:1", Load: true}
	out := buildAndCapture(t, "bake", "", spec)
	if !strings.Contains(out, `"type=docker"`) {
		t.Errorf("load should map to type=docker output: %s", out)
	}
}

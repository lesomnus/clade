// Package builder runs container image builds. Each build strategy registers
// itself by kind and is constructed from the raw `build` config of a port plus
// a Spec describing what to produce. A constructed Builder carries everything it
// needs, so the interface is just Build(ctx).
package builder

import (
	"context"
	"fmt"
	"io"
)

// Spec is the runtime description of a single build: what image to produce and
// from what base. It is independent of the build strategy; strategy-specific
// options come from the port's raw build config instead.
type Spec struct {
	// Dir is the port directory; relative paths resolve against it.
	Dir string
	// Tags are the full references to tag the image with, e.g. "repo:1.2.3".
	Tags []string
	// Base is the upstream image reference, injected as the BASE build arg.
	Base string
	// Labels are labels to inject (e.g. the base name and digest).
	Labels map[string]string

	// Push pushes the result to the registry; Load loads it into the local
	// image store. They are mutually exclusive.
	Push bool
	Load bool

	// DryRun prints the command instead of executing it.
	DryRun bool
	// Bin is the binary to invoke (default "docker").
	Bin string
	// Stdout and Stderr receive command output (default os.Stdout/os.Stderr).
	Stdout io.Writer
	Stderr io.Writer
}

// Builder performs a single, fully-configured build.
type Builder interface {
	Build(ctx context.Context) error
}

// Factory constructs a Builder from a port's raw build config and a Spec.
type Factory func(params []byte, spec Spec) (Builder, error)

var factories = map[string]Factory{}

// Register makes a build strategy available under kind. It panics on a
// duplicate registration and is intended to be called from init.
func Register(kind string, f Factory) {
	if _, dup := factories[kind]; dup {
		panic(fmt.Sprintf("builder: kind %q already registered", kind))
	}
	factories[kind] = f
}

// New constructs the Builder registered under kind (empty means "build").
func New(kind string, params []byte, spec Spec) (Builder, error) {
	if kind == "" {
		kind = "build"
	}
	f, ok := factories[kind]
	if !ok {
		return nil, fmt.Errorf("builder: unknown kind %q", kind)
	}
	return f(params, spec)
}

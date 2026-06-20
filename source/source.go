// Package source abstracts where the upstream versions of a port come from. Not
// every upstream is a container registry, so the discovery strategy is
// pluggable: implementations register by kind and are constructed from the raw
// source spec declared in a port.yaml.
package source

import (
	"context"
	"fmt"
)

// Source lists the candidate upstream versions to track. The returned strings
// are fed to a tag.Selector, which parses and selects among them.
type Source interface {
	Versions(ctx context.Context) ([]string, error)
}

// Deps carries collaborators a source may need. A tags lister is injected so the
// source package need not depend on the registry client.
type Deps struct {
	// Tags lists the tags of an OCI repository. The container source uses it.
	Tags func(ctx context.Context, repo string) ([]string, error)
}

// Factory builds a Source from the raw YAML of a source spec and its deps.
type Factory func(params []byte, deps Deps) (Source, error)

var registry = map[string]Factory{}

// Register makes a discovery strategy available under kind. It panics on a
// duplicate registration and is intended to be called from init.
func Register(kind string, f Factory) {
	if _, dup := registry[kind]; dup {
		panic(fmt.Sprintf("source: kind %q already registered", kind))
	}
	registry[kind] = f
}

// New constructs the Source registered under kind, decoding params.
func New(kind string, params []byte, deps Deps) (Source, error) {
	f, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("source: unknown kind %q", kind)
	}
	return f(params, deps)
}

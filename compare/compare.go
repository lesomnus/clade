// Package compare abstracts how clade decides whether a built image is outdated
// with respect to its base image. Different strategies (creation time, base
// digest, and potentially labels in the future) are pluggable: implementations
// register by kind and are constructed from a raw config blob.
//
// A comparator only ever sees two existing images. A missing target image is
// treated as outdated by the caller before any comparator runs.
package compare

import (
	"context"
	"fmt"

	"github.com/lesomnus/clade/registry"
)

// Comparator reports whether target is outdated relative to its base.
type Comparator interface {
	IsOutdated(ctx context.Context, base, target *registry.ImageInfo) (bool, error)
}

// Factory builds a Comparator from a raw YAML config blob (may be empty).
type Factory func(params []byte) (Comparator, error)

var factories = map[string]Factory{}

// Register makes a strategy available under kind. It panics on a duplicate
// registration and is intended to be called from init.
func Register(kind string, f Factory) {
	if _, dup := factories[kind]; dup {
		panic(fmt.Sprintf("compare: kind %q already registered", kind))
	}
	factories[kind] = f
}

// New constructs the Comparator registered under kind, decoding params.
func New(kind string, params []byte) (Comparator, error) {
	f, ok := factories[kind]
	if !ok {
		return nil, fmt.Errorf("compare: unknown kind %q", kind)
	}
	return f(params)
}

// Package tag abstracts selection of upstream tags. Not every tagging scheme is
// semver, so the selection strategy is pluggable: implementations register
// themselves by kind and are constructed from the raw target spec declared in a
// port.yaml.
package tag

import "fmt"

// Matched is a selected upstream tag together with the data used to render the
// build tag template. For the semver strategy, Data is a *semver.Version, whose
// Major/Minor/Patch methods are available to the template.
type Matched struct {
	Tag  string
	Data any
}

// Selector chooses which of the available upstream tags to track.
type Selector interface {
	Select(tags []string) ([]Matched, error)
}

// Factory builds a Selector from the raw YAML of a target spec.
type Factory func(params []byte) (Selector, error)

var registry = map[string]Factory{}

// Register makes a selection strategy available under kind. It panics on a
// duplicate registration and is intended to be called from init.
func Register(kind string, f Factory) {
	if _, dup := registry[kind]; dup {
		panic(fmt.Sprintf("tag: kind %q already registered", kind))
	}
	registry[kind] = f
}

// New constructs the Selector registered under kind, decoding params.
func New(kind string, params []byte) (Selector, error) {
	f, ok := registry[kind]
	if !ok {
		return nil, fmt.Errorf("tag: unknown kind %q", kind)
	}
	return f(params)
}

// Package port describes a buildable image ("port") and parses its port.yaml.
//
// A port lives in a directory together with its Dockerfile and build context:
//
//	ports/
//	  golang-dev/
//	    Dockerfile
//	    port.yaml
//
// port.yaml declares the upstream image to track and how the built image is
// tagged and pushed:
//
//	parent:
//	  repo: docker.io/library/golang
//	  target:
//	    kind: semver
//	    last-major: 2
//	    last-minor: 3
//	    match: "-alpine$"
//	build:
//	  repo: my-registry/my-image/golang-dev
//	  tag: "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine"
package port

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// Port is a single buildable image definition.
type Port struct {
	// Dir is the directory that holds the Dockerfile, context and port.yaml.
	// It is set on load and is not read from the YAML document.
	Dir string `yaml:"-"`

	Parent Parent `yaml:"parent"`
	Build  Build  `yaml:"build"`
}

// Parent declares the upstream image to track and how its tags are selected.
type Parent struct {
	// Repo is the upstream repository, e.g. "docker.io/library/golang".
	// It may also be the Build.Repo of another port, forming an internal edge.
	Repo   string `yaml:"repo"`
	Target Target `yaml:"target"`
}

// Build declares how the produced image is named.
type Build struct {
	// Repo is the destination repository to push to.
	Repo string `yaml:"repo"`
	// Tag is a Go text/template rendered with the data of each selected
	// upstream tag, e.g. "{{.Major}}.{{.Minor}}.{{.Patch}}-alpine".
	Tag string `yaml:"tag"`
}

// Target is the tag selection spec. The concrete selection strategy is
// abstracted behind Kind; the remaining fields are strategy specific and kept
// as raw YAML so that the tag package can decode them without this package
// depending on it.
type Target struct {
	// Kind names the selection strategy, e.g. "semver".
	Kind string
	// Params is the raw YAML of the whole target mapping (including kind).
	Params []byte
}

// UnmarshalYAML implements goccy/go-yaml's BytesUnmarshaler. It records the raw
// node so kind-specific parameters can be decoded later, and extracts Kind.
func (t *Target) UnmarshalYAML(b []byte) error {
	var head struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(b, &head); err != nil {
		return fmt.Errorf("decode target: %w", err)
	}

	t.Kind = head.Kind
	t.Params = b
	return nil
}

// Validate reports whether the port is well formed.
func (p *Port) Validate() error {
	switch {
	case p.Parent.Repo == "":
		return fmt.Errorf("parent.repo is required")
	case p.Parent.Target.Kind == "":
		return fmt.Errorf("parent.target.kind is required")
	case p.Build.Repo == "":
		return fmt.Errorf("build.repo is required")
	case p.Build.Tag == "":
		return fmt.Errorf("build.tag is required")
	}
	return nil
}

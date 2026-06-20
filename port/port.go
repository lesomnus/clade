// Package port describes a buildable image ("port") and parses its port.yaml.
//
// A port lives in a directory together with its Dockerfile and build context:
//
//	ports/
//	  dev-golang/
//	    Dockerfile
//	    port.yaml
//
// port.yaml declares the upstream source to track, how its versions are
// selected, how the built image is named and pushed, and (optionally) how
// outdatedness is judged:
//
//	source:
//	  kind: container
//	  repo: docker.io/library/golang
//	select:
//	  kind: semver
//	  last-major: 2
//	  last-minor: 3
//	build:
//	  repo: my-registry/my-image/dev-golang
//	  tags:
//	    - "{{.Major}}.{{.Minor}}.{{.Patch}}"
//	    - "{{.Major}}.{{.Minor}}"
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

	Source  Source        `yaml:"source"`
	Select  Select        `yaml:"select"`
	Compare []CompareSpec `yaml:"compare"`
	Build   Build         `yaml:"build"`
}

// Source declares where the upstream versions to track come from. Kind selects
// the discovery strategy ("container" lists an OCI repository's tags, "http"
// fetches a version string from a URL); the remaining fields are strategy
// specific. Repo and Url are extracted for the graph (internal-edge detection
// and the base reference); Params keeps the raw node for the source package.
type Source struct {
	// Kind names the discovery strategy, e.g. "container" or "http".
	Kind string
	// Repo is the upstream OCI repository for kind "container". It may also be
	// the Build.Repo of another port, forming an internal edge.
	Repo string
	// Url is the endpoint for kind "http".
	Url string
	// Params is the raw YAML of the whole source mapping (including kind).
	Params []byte
}

// UnmarshalYAML implements goccy/go-yaml's BytesUnmarshaler. It extracts the
// fields the graph needs and keeps the raw node for the source package.
func (s *Source) UnmarshalYAML(b []byte) error {
	var head struct {
		Kind string `yaml:"kind"`
		Repo string `yaml:"repo"`
		Url  string `yaml:"url"`
	}
	if err := yaml.Unmarshal(b, &head); err != nil {
		return fmt.Errorf("decode source: %w", err)
	}

	s.Kind = head.Kind
	s.Repo = head.Repo
	s.Url = head.Url
	s.Params = b
	return nil
}

// Build declares how the produced image is named and which build strategy
// produces it. Strategy-specific options (Dockerfile, context, platforms,
// cache, ...) are kept as raw YAML in Params and decoded by the builder of the
// selected Kind, so this package stays independent of any build backend.
type Build struct {
	// Repo is the destination repository to push to.
	Repo string
	// Tags are Go text/templates rendered with the data of each selected
	// upstream version, e.g. "{{.Major}}.{{.Minor}}.{{.Patch}}". A single built
	// image is tagged with every rendered tag; the first is canonical (its
	// absence marks the node outdated) and the rest are typically floating tags
	// (e.g. "{{.Major}}.{{.Minor}}").
	Tags []string
	// Kind names the build strategy, e.g. "build" (docker buildx build) or
	// "bake" (docker buildx bake). Empty means the default.
	Kind string
	// Params is the raw YAML of the whole build mapping (including the fields
	// above), passed to the builder factory.
	Params []byte
}

// UnmarshalYAML implements goccy/go-yaml's BytesUnmarshaler. It extracts the
// fields needed to build the graph and keeps the raw node for the builder.
func (b *Build) UnmarshalYAML(data []byte) error {
	var head struct {
		Repo string   `yaml:"repo"`
		Tags []string `yaml:"tags"`
		Kind string   `yaml:"kind"`
	}
	if err := yaml.Unmarshal(data, &head); err != nil {
		return fmt.Errorf("decode build: %w", err)
	}

	b.Repo = head.Repo
	b.Tags = head.Tags
	b.Kind = head.Kind
	b.Params = data
	return nil
}

// Select is the version selection spec. The concrete selection strategy is
// abstracted behind Kind; the remaining fields are strategy specific and kept
// as raw YAML so that the tag package can decode them without this package
// depending on it.
type Select struct {
	// Kind names the selection strategy, e.g. "semver".
	Kind string
	// Params is the raw YAML of the whole select mapping (including kind).
	Params []byte
}

// UnmarshalYAML implements goccy/go-yaml's BytesUnmarshaler. It records the raw
// node so kind-specific parameters can be decoded later, and extracts Kind.
func (s *Select) UnmarshalYAML(b []byte) error {
	var head struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(b, &head); err != nil {
		return fmt.Errorf("decode select: %w", err)
	}

	s.Kind = head.Kind
	s.Params = b
	return nil
}

// CompareSpec is one entry of a port's compare chain: a strategy kind plus its
// raw params. The chain is tried in order with fallback (see package compare).
// An empty list means the per-source-kind default is used.
type CompareSpec struct {
	// Kind names the comparison strategy, e.g. "created" or "digest".
	Kind string
	// Params is the raw YAML of the whole compare entry (including kind).
	Params []byte
}

// UnmarshalYAML implements goccy/go-yaml's BytesUnmarshaler.
func (c *CompareSpec) UnmarshalYAML(b []byte) error {
	var head struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(b, &head); err != nil {
		return fmt.Errorf("decode compare: %w", err)
	}

	c.Kind = head.Kind
	c.Params = b
	return nil
}

// Validate reports whether the port is well formed.
func (p *Port) Validate() error {
	switch {
	case p.Source.Kind == "":
		return fmt.Errorf("source.kind is required")
	case p.Source.Kind == "container" && p.Source.Repo == "":
		return fmt.Errorf("source.repo is required for kind \"container\"")
	case p.Source.Kind == "http" && p.Source.Url == "":
		return fmt.Errorf("source.url is required for kind \"http\"")
	case p.Select.Kind == "":
		return fmt.Errorf("select.kind is required")
	case p.Build.Repo == "":
		return fmt.Errorf("build.repo is required")
	case len(p.Build.Tags) == 0:
		return fmt.Errorf("build.tags is required")
	}
	for i, t := range p.Build.Tags {
		if t == "" {
			return fmt.Errorf("build.tags[%d] is empty", i)
		}
	}
	for i, c := range p.Compare {
		if c.Kind == "" {
			return fmt.Errorf("compare[%d].kind is required", i)
		}
	}
	return nil
}

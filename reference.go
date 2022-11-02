package clade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/pl"
	"gopkg.in/yaml.v3"
)

type RefNamedPipelineTagged interface {
	reference.NamedTagged
	Pipeline() *pl.Pl
}

type refNamedPipelineTagged struct {
	reference.Named
	tag      string
	pipeline *pl.Pl
}

func (r *refNamedPipelineTagged) Domain() string {
	return reference.Domain(r.Named)
}

func (r *refNamedPipelineTagged) Path() string {
	return reference.Path(r.Named)
}

func (r *refNamedPipelineTagged) Tag() string {
	return r.tag
}

func (r *refNamedPipelineTagged) String() string {
	return fmt.Sprintf("%s:%s", r.Name(), r.tag)
}

func (r *refNamedPipelineTagged) Pipeline() *pl.Pl {
	return r.pipeline
}

func (r *refNamedPipelineTagged) UnmarshalYAML(n *yaml.Node) error {
	ref_str := ""

	switch n.Kind {
	case yaml.ScalarNode:
		if err := n.Decode(&ref_str); err != nil {
			return err
		}

	case yaml.MappingNode:
		type refMap struct {
			Name string
			Tag  string
		}

		var ref refMap
		if err := n.Decode(&ref); err != nil {
			return err
		}

		ref_str = fmt.Sprintf("%s:%s", ref.Name, ref.Tag)

	default:
		return errors.New("invalid node type")
	}

	ref, err := ParseRefNamedTagged(ref_str)
	if err != nil {
		return err
	}

	*r = *asRefNamedPipelineTagged(ref)

	return nil
}

// ParseRefNamedTagged parses given string into named tagged reference.
// If the tag is a pipeline expression, returned reference implements `RefNamedPipelineTagged`.
func ParseRefNamedTagged(s string) (reference.NamedTagged, error) {
	pos := strings.LastIndex(s, ":")
	if pos < 0 {
		return nil, errors.New("no tag")
	}

	tag := s[pos+1:]
	if tag == "" {
		return nil, errors.New("no tag")
	}

	named, err := reference.ParseNamed(s[:pos])
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(tag, "(") && strings.HasSuffix(tag, ")") {
		pl, err := pl.ParseString(tag)
		if err != nil {
			return nil, fmt.Errorf("invalid pipeline: %w", err)
		}

		return &refNamedPipelineTagged{
			Named:    named,
			tag:      tag,
			pipeline: pl,
		}, nil
	} else {
		return reference.WithTag(named, tag)
	}
}

func asRefNamedPipelineTagged(ref reference.NamedTagged) *refNamedPipelineTagged {
	pl_ref, ok := ref.(*refNamedPipelineTagged)
	if ok {
		return pl_ref
	}

	fn, _ := pl.NewFn("pass", ref.Tag())

	return &refNamedPipelineTagged{
		Named:    ref,
		tag:      ref.Tag(),
		pipeline: pl.NewPl(fn),
	}
}

// AsRefNamedPipelineTagged returns reference implements RefNamedPipelineTagged by creating
// pass-only pipeline that passes tag string of given reference.
func AsRefNamedPipelineTagged(ref reference.NamedTagged) RefNamedPipelineTagged {
	return asRefNamedPipelineTagged(ref)
}

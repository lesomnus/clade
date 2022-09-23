package clade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/pipeline"
	"gopkg.in/yaml.v3"
)

type RefNamedPipelineTagged interface {
	reference.NamedTagged
	Pipeline() pipeline.Pipeline
}

type refNamedPipelineTagged struct {
	reference.Named
	tag      string
	pipeline pipeline.Pipeline
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

func (r *refNamedPipelineTagged) Pipeline() pipeline.Pipeline {
	return r.pipeline
}

func (r *refNamedPipelineTagged) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		var s string
		if err := n.Decode(&s); err != nil {
			return err
		}

		ref, err := ParseReference(s)
		if err != nil {
			return err
		} else {
			r.Named = ref
		}

		if tagged, ok := ref.(reference.Tagged); !ok {
			return errors.New("reference must be tagged")
		} else {
			r.tag = tagged.Tag()
		}

		if pled, ok := ref.(RefNamedPipelineTagged); ok {
			r.pipeline = pled.Pipeline()
		} else {
			r.pipeline = pipeline.Pipeline{&pipeline.Fn{Name: ">", Args: []any{r.tag}}}
		}

	case yaml.MappingNode:
		type refMap struct {
			Name string
			Tag  *tagExpr
		}

		var ref refMap
		if err := n.Decode(&ref); err != nil {
			return err
		}

		if named, err := reference.ParseNamed(ref.Name); err != nil {
			return err
		} else {
			r.Named = named
		}

		r.tag = ref.Tag.Tag
		r.pipeline = ref.Tag.Pl

	default:
		return errors.New("invalid node type")
	}

	return nil
}

type tagExpr struct {
	Tag string
	Pl  pipeline.Pipeline
}

func (r *tagExpr) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		if err := n.Decode(&r.Tag); err != nil {
			return err
		}

		expr := "x.x/x/x:" + r.Tag
		var ref refNamedPipelineTagged
		if err := yaml.Unmarshal([]byte(expr), &ref); err != nil {
			return err
		}

		r.Pl = ref.pipeline

	case yaml.SequenceNode:
		if err := n.Decode(&r.Pl); err != nil {
			return err
		}

	default:
		return errors.New("invalid node type")
	}

	return nil
}

func ParseReference(s string) (reference.Named, error) {
	ref, err := reference.ParseNamed(s)
	if err != nil {
		if !errors.Is(err, reference.ErrReferenceInvalidFormat) {
			return nil, err
		}

		pos := strings.LastIndex(s, ":")
		if pos < 0 {
			// Not tag field
			return nil, err
		}

		named, err := reference.ParseNamed(s[:pos])
		if err != nil {
			return nil, err
		}

		tag := s[pos+1:]
		if strings.HasPrefix(tag, "{") && strings.HasSuffix(tag, "}") {
			pl, err := pipeline.Parse(tag[1 : len(tag)-1])
			if err != nil {
				return nil, fmt.Errorf("%w: %s", reference.ErrDigestInvalidFormat, err.Error())
			}

			ref = &refNamedPipelineTagged{
				Named:    named,
				tag:      tag,
				pipeline: pl,
			}
		} else {
			return nil, reference.ErrTagInvalidFormat
		}
	}

	return ref, nil
}

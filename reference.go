package clade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/pipeline"
)

type refNamedTagged struct {
	reference.Named
	tag string
}

func (r *refNamedTagged) Domain() string {
	return reference.Domain(r.Named)
}

func (r *refNamedTagged) Path() string {
	return reference.Path(r.Named)
}

func (r *refNamedTagged) Tag() string {
	return r.tag
}

func (r *refNamedTagged) String() string {
	return fmt.Sprintf("%s:%s", r.Name(), r.tag)
}

type RefNamedPipelineTagged interface {
	reference.NamedTagged
	Pipeline() pipeline.Pipeline
}

type refNamedPipelineTagged struct {
	refNamedTagged
	pipeline pipeline.Pipeline
}

func (r *refNamedPipelineTagged) Pipeline() pipeline.Pipeline {
	return r.pipeline
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
				refNamedTagged: refNamedTagged{
					Named: named,
					tag:   tag,
				},
				pipeline: pl,
			}
		} else {
			return nil, reference.ErrTagInvalidFormat
		}
	}

	return ref, nil
}

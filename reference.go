package clade

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/distribution/distribution/reference"
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

type RefNamedRegexTagged interface {
	reference.NamedTagged
	Pattern() *regexp.Regexp
}

type refNamedRegexTagged struct {
	refNamedTagged
	pattern *regexp.Regexp
}

func (r *refNamedRegexTagged) Pattern() *regexp.Regexp {
	return r.pattern
}

type RefNamedPipelineTagged interface {
	reference.NamedTagged
	PipelineExpr() string
}

type refNamedPipelineTagged struct {
	refNamedTagged
}

func (r *refNamedPipelineTagged) PipelineExpr() string {
	return fmt.Sprintf("{%s}", r.tag)
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
		if strings.HasPrefix(tag, "/") && strings.HasSuffix(tag, "/") {
			pattern, err := regexp.Compile(tag)
			if err != nil {
				return nil, reference.ErrTagInvalidFormat
			}

			ref = &refNamedRegexTagged{
				refNamedTagged: refNamedTagged{
					Named: named,
					tag:   tag,
				},
				pattern: pattern,
			}
		} else if strings.HasPrefix(tag, "{") && strings.HasSuffix(tag, "}") {
			ref = &refNamedPipelineTagged{
				refNamedTagged: refNamedTagged{
					Named: named,
					tag:   tag,
				},
			}
		} else {
			return nil, reference.ErrTagInvalidFormat
		}
	}

	return ref, nil
}

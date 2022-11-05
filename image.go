package clade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/pl"
	"gopkg.in/yaml.v3"
)

type Pipeliner interface {
	String() string
	Pipeline() *pl.Pl
}

type pipeliner struct {
	expr string
	pl   *pl.Pl
}

func (p *pipeliner) String() string {
	return p.expr
}

func (p *pipeliner) Pipeline() *pl.Pl {
	return p.pl
}

func (p *pipeliner) UnmarshalYAML(n *yaml.Node) error {
	if err := n.Decode(&p.expr); err != nil {
		return err
	}

	if strings.HasPrefix(p.expr, "(") && strings.HasSuffix(p.expr, ")") {
		pl, err := pl.ParseString(p.expr)
		if err != nil {
			return fmt.Errorf("invalid pipeline expression: %w", err)
		}
		p.pl = pl
	} else {
		fn, _ := pl.NewFn("pass", p.expr)
		p.pl = pl.NewPl(fn)
	}

	return nil
}

type Image struct {
	reference.Named `yaml:"-"`

	Tags []Pipeliner            `yaml:"-"`
	From RefNamedPipelineTagged `yaml:"-"`
	Args map[string]string

	Dockerfile  string
	ContextPath string `yaml:"context"`
}

func (i *Image) UnmarshalYAML(n *yaml.Node) error {
	type Image_ Image
	if err := n.Decode((*Image_)(i)); err != nil {
		return err
	}

	type I struct {
		Tags []*pipeliner
		From *refNamedPipelineTagged
	}
	var tmp I
	if err := n.Decode(&tmp); err != nil {
		return err
	}

	i.Tags = make([]Pipeliner, len(tmp.Tags))
	for j, tag := range tmp.Tags {
		i.Tags[j] = tag
	}

	i.From = tmp.From

	return nil
}

type ResolvedImage struct {
	reference.Named

	Tags []string
	From reference.NamedTagged
	Args map[string]string

	Dockerfile  string
	ContextPath string
}

func (i *ResolvedImage) Tagged() (reference.NamedTagged, error) {
	if len(i.Tags) == 0 {
		return nil, errors.New("not tagged")
	}

	tagged, err := reference.WithTag(i.Named, i.Tags[0])
	if err != nil {
		return nil, err
	}

	return tagged, nil
}

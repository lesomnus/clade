package clade

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/distribution/distribution/reference"
	ba "github.com/lesomnus/boolal"
	"github.com/lesomnus/pl"
	"gopkg.in/yaml.v3"
)

const AnnotationDerefId = "clade.deref.id"

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

	Skip bool                   `yaml:"skip"`
	Tags []Pipeliner            `yaml:"-"`
	From RefNamedPipelineTagged `yaml:"-"`
	Args map[string]Pipeliner   `yaml:"-"`

	Dockerfile  string
	ContextPath string   `yaml:"context"`
	Platform    *ba.Expr `yaml:"-"`
}

func (i *Image) UnmarshalYAML(n *yaml.Node) error {
	type Image_ Image
	if err := n.Decode((*Image_)(i)); err != nil {
		return err
	}

	type I struct {
		Tags []*pipeliner
		From *refNamedPipelineTagged
		Args map[string]*pipeliner

		Platform string
	}
	var tmp I
	if err := n.Decode(&tmp); err != nil {
		return err
	}

	if tmp.Platform == "" {
		i.Platform = nil
	} else if expr, err := ba.ParseString(tmp.Platform); err != nil {
		return fmt.Errorf("platform: %w", err)
	} else {
		i.Platform = expr
	}

	i.Tags = make([]Pipeliner, len(tmp.Tags))
	for j, tag := range tmp.Tags {
		i.Tags[j] = tag
	}

	i.From = tmp.From

	i.Args = make(map[string]Pipeliner, len(tmp.Args))
	for key, arg := range tmp.Args {
		i.Args[key] = arg
	}

	return nil
}

type ResolvedImage struct {
	reference.Named

	Skip bool
	Tags []string
	From reference.NamedTagged
	Args map[string]string

	Dockerfile  string
	ContextPath string
	Platform    *ba.Expr
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

func CalcDerefId(dgsts ...[]byte) string {
	max_len := 0
	for _, dgst := range dgsts {
		if max_len < len(dgst) {
			max_len = len(dgst)
		}
	}

	rst := make([]byte, max_len)
	for _, dgst := range dgsts {
		for i, v := range dgst {
			rst[i] = rst[i] ^ v
		}
	}

	return hex.EncodeToString(rst)
}

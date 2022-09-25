package clade

import (
	"errors"

	"github.com/distribution/distribution/reference"
	"gopkg.in/yaml.v3"
)

type Image struct {
	reference.Named `yaml:"-"`

	Tags []string
	From RefNamedPipelineTagged `yaml:"-"`
	Args map[string]string

	Dockerfile  string
	ContextPath string
}

func (i *Image) UnmarshalYAML(n *yaml.Node) error {
	type Image_ Image
	if err := n.Decode((*Image_)(i)); err != nil {
		return err
	}

	type I struct{ From *refNamedPipelineTagged }
	var tmp I
	if err := n.Decode(&tmp); err != nil {
		return err
	}

	i.From = tmp.From

	return nil
}

func (i *Image) Tagged() (reference.NamedTagged, error) {
	if tagged, ok := i.Named.(reference.NamedTagged); ok {
		return tagged, nil
	}

	if len(i.Tags) == 0 {
		return nil, errors.New("not tagged")
	}

	tagged, err := reference.WithTag(i.Named, i.Tags[0])
	if err != nil {
		return nil, err
	}

	return tagged, nil
}

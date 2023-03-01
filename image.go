package clade

import (
	"encoding/hex"
	"errors"

	"github.com/distribution/distribution/v3/reference"
	"gopkg.in/yaml.v3"
)

const AnnotationDerefId = "clade.deref.id"

type BaseImage struct {
	Primary     *ImageReference
	Secondaries []*ImageReference
}

func (i *BaseImage) unmarshalYamlScalar(node *yaml.Node) error {
	i.Secondaries = nil
	return node.Decode(&i.Primary)
}

func (i *BaseImage) unmarshalYamlMap(node *yaml.Node) error {
	var ref struct {
		Name string
		Tags string
		With []*ImageReference
	}
	if err := node.Decode(&ref); err != nil {
		return err
	}

	i.Secondaries = ref.With
	i.Primary = &ImageReference{}
	if err := i.Primary.FromNameTag(ref.Name, ref.Tags); err != nil {
		return err
	}

	return nil
}

func (i *BaseImage) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		return i.unmarshalYamlScalar(node)

	case yaml.MappingNode:
		return i.unmarshalYamlMap(node)
	}

	return &yaml.TypeError{Errors: []string{"must be string or map"}}
}

type Image struct {
	reference.Named `yaml:"-"`

	Skip bool                `yaml:"skip"`
	Tags []Pipeline          `yaml:"tags"`
	From BaseImage           `yaml:"from"`
	Args map[string]Pipeline `yaml:"args"`

	Dockerfile  string       `yaml:"dockerfile"`
	ContextPath string       `yaml:"context"`
	Platform    *BoolAlgebra `yaml:"platform"`
}

type ResolvedBaseImage struct {
	Primary     reference.NamedTagged
	Secondaries []reference.NamedTagged
}

type ResolvedImage struct {
	reference.Named

	Skip bool
	Tags []string
	From *ResolvedBaseImage
	Args map[string]string

	Dockerfile  string
	ContextPath string
	Platform    *BoolAlgebra
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

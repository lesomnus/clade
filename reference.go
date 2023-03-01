package clade

import (
	"fmt"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"
	"gopkg.in/yaml.v3"
)

type ImageReference struct {
	reference.Named
	Tag *Pipeline
}

func (r *ImageReference) FromNameTag(name string, tag string) error {
	named, err := reference.ParseNamed(name)
	if err != nil {
		return fmt.Errorf("parse reference name: %w", err)
	} else {
		r.Named = named
	}

	if strings.HasPrefix(tag, "(") && strings.HasSuffix(tag, ")") {
		// Pipeline expression
	} else if strings.ContainsRune(tag, ':') {
		if _, err := reference.WithDigest(named, digest.Digest(tag)); err != nil {
			return err
		}
	} else {
		if _, err := reference.WithTag(named, tag); err != nil {
			return err
		}
	}

	if err := yaml.Unmarshal([]byte(tag), &r.Tag); err != nil {
		return fmt.Errorf("unmarshal reference tag: %w", err)
	}

	return nil
}

func (r *ImageReference) unmarshalYamlScalar(node *yaml.Node) error {
	ref := ""
	if err := node.Decode(&ref); err != nil {
		panic(err)
	}

	pos := strings.LastIndex(ref, "/") + 1
	pos += strings.IndexAny(ref[pos:], ":@")
	if (ref[pos] == '@') && !strings.ContainsRune(ref[pos+1:], ':') {
		return reference.ErrDigestInvalidFormat
	}
	if err := r.FromNameTag(ref[:pos], ref[pos+1:]); err != nil {
		return err
	}

	return nil
}

func (r *ImageReference) unmarshalYamlMap(node *yaml.Node) error {
	var ref struct {
		Name string
		Tag  string
	}
	if err := node.Decode(&ref); err != nil {
		panic(err)
	}
	if err := r.FromNameTag(ref.Name, ref.Tag); err != nil {
		return err
	}

	return nil
}

func (r *ImageReference) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		return r.unmarshalYamlScalar(node)

	case yaml.MappingNode:
		return r.unmarshalYamlMap(node)
	}

	return &yaml.TypeError{Errors: []string{"must be string or map"}}
}

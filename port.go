package clade

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/sv"
	"github.com/lesomnus/pl"
	"github.com/rs/zerolog"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type Port struct {
	Name reference.Named   `yaml:"-"`
	Args map[string]string `yaml:"args"`

	Skip        bool         `yaml:"skip"`
	Dockerfile  string       `yaml:"dockerfile"`
	ContextPath string       `yaml:"context"`
	Platform    *BoolAlgebra `yaml:"platform"`

	Images []*Image
}

func (p *Port) UnmarshalYAML(n *yaml.Node) error {
	type Port_ Port
	if err := n.Decode((*Port_)(p)); err != nil {
		return err
	}

	var tmp struct{ Name string }
	if err := n.Decode(&tmp); err != nil {
		return err
	}

	named, err := reference.ParseNamed(tmp.Name)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	for _, image := range p.Images {
		if image == nil {
			continue
		}

		image.Named = named

		if p.Skip {
			image.Skip = p.Skip
		}

		if image.Platform == nil {
			image.Platform = p.Platform
		}

		if image.Args == nil {
			image.Args = make(map[string]Pipeline)
		}
		for k, v := range p.Args {
			if _, ok := image.Args[k]; ok {
				continue
			}

			fn, _ := pl.NewFn("pass", v)
			image.Args[k] = Pipeline(*pl.NewPl(fn))
		}
	}

	p.Name = named
	return nil
}

func DeduplicateBySemver(lhs *[]string, rhs *[]string) error {
	highest := func(vs []string) (semver.Version, error) {
		rst := semver.Version{}
		for _, v := range vs {
			v, err := sv.Parse(v)
			if err != nil {
				return rst, fmt.Errorf("failed to parse semver: %w", err)
			}

			if rst.LT(v.Version) {
				rst = v.Version
			}
		}

		return rst, nil
	}

	vl, err := highest(*lhs)
	if err != nil {
		return err
	}

	vr, err := highest(*rhs)
	if err != nil {
		return err
	}

	if vl.EQ(vr) {
		*rhs = (*rhs)[0:0]
		return nil
	}

	major, minor := lhs, rhs
	if vl.LT(vr) {
		major, minor = rhs, lhs
	}

	cursor := 0
	for _, tag := range *minor {
		if slices.Contains(*major, tag) {
			continue
		}

		(*minor)[cursor] = tag
		cursor++
	}

	(*minor) = (*minor)[0:cursor]

	return nil
}

// ReadPort parses port file at given path.
// It fills empty fields in children by using fields in parent if possible according to Port rules.
// For example, if .Images[].Dockerfile empty, it will use .Dockerfile.
func ReadPort(path string) (*Port, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dirname := filepath.Dir(path)

	port := &Port{}
	if err := yaml.Unmarshal(data, &port); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	if p, err := ResolvePath(dirname, port.Dockerfile, "Dockerfile"); err != nil {
		return nil, fmt.Errorf("failed to resolve path to Dockerfile: %w", err)
	} else {
		port.Dockerfile = p
	}

	if p, err := ResolvePath(dirname, port.ContextPath, "."); err != nil {
		return nil, fmt.Errorf("failed to resolve path to context: %w", err)
	} else {
		port.ContextPath = p
	}

	for _, image := range port.Images {
		if p, err := ResolvePath(dirname, image.Dockerfile, port.Dockerfile); err != nil {
			return nil, fmt.Errorf("failed to resolve path to Dockerfile: %w", err)
		} else {
			image.Dockerfile = p
		}

		if p, err := ResolvePath(dirname, image.ContextPath, port.ContextPath); err != nil {
			return nil, fmt.Errorf("failed to resolve path to context: %w", err)
		} else {
			image.ContextPath = p
		}
	}

	return port, nil
}

func ReadPortsFromFs(ctx context.Context, path string) ([]*Port, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Str("path", path).Msg("read ports")

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	ports := make([]*Port, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		port_path := filepath.Join(path, entry.Name(), "port.yaml")
		l.Debug().Str("path", port_path).Msg("read port")

		port, err := ReadPort(port_path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("failed to read port at %s: %w", port_path, err)
		}

		ports = append(ports, port)
	}

	return ports, nil
}

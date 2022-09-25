package clade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/blang/semver/v4"
	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/sv"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

type Port struct {
	Name reference.Named `yaml:"-"`
	Args map[string]string

	Dockerfile  string
	ContextPath string

	Images []*Image
}

func (p *Port) UnmarshalYAML(n *yaml.Node) error {
	type Port_ Port
	if err := n.Decode((*Port_)(p)); err != nil {
		return err
	}

	type P struct{ Name string }
	var tmp P
	if err := n.Decode(&tmp); err != nil {
		return err
	}

	named, err := reference.ParseNamed(tmp.Name)
	if err != nil {
		return fmt.Errorf("name: %w", err)
	}

	for i := range p.Images {
		p.Images[i].Named = named
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

// ResolvePath returns an absolute representation of joined path of `base` and `path`.
// If the joined path is not absolute, it will be joined with the current working directory to turn it into an absolute path.
// If the `path` is empty, the `base` is joined with `fallback`.
func ResolvePath(base string, path string, fallback string) (string, error) {
	if base == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		base = wd
	}

	if !filepath.IsAbs(base) {
		abs, err := filepath.Abs(base)
		if err != nil {
			return "", err
		}

		base = abs
	}

	if path == "" {
		path = fallback
	}

	if filepath.IsAbs(path) {
		return path, nil
	}

	return filepath.Join(base, path), nil
}

// ReadPort parses port file at given path.
// If fills empty fields in children by using fields in parent if possible according to Port rules.
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

	for i, image := range port.Images {
		if p, err := ResolvePath(dirname, image.Dockerfile, port.Dockerfile); err != nil {
			return nil, fmt.Errorf("failed to resolve path to Dockerfile: %w", err)
		} else {
			port.Images[i].Dockerfile = p
		}

		if p, err := ResolvePath(dirname, image.ContextPath, port.ContextPath); err != nil {
			return nil, fmt.Errorf("failed to resolve path to context: %w", err)
		} else {
			port.Images[i].ContextPath = p
		}

		if image.Args == nil {
			image.Args = make(map[string]string)
		}
		for k, v := range port.Args {
			if _, ok := image.Args[k]; ok {
				continue
			}

			image.Args[k] = v
		}
	}

	return port, nil
}

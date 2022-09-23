package clade

import (
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/sv"
	"golang.org/x/exp/slices"
)

type Port struct {
	Name string
	Args map[string]string

	Dockerfile  string
	ContextPath string

	Images []Image
}

func (p *Port) ParseImages() ([]*NamedImage, error) {
	name, err := reference.ParseNamed(p.Name)
	if err != nil {
		return nil, err
	}

	imgs := make([]*NamedImage, len(p.Images))
	for i, img := range p.Images {
		named_img := &NamedImage{
			Named: name,
			Tags:  slices.Clone(img.Tags),
			From:  img.From,
			Args:  make(map[string]string, len(p.Args)+len(img.Args)),

			Dockerfile:  img.Dockerfile,
			ContextPath: img.ContextPath,
		}

		for k, v := range p.Args {
			named_img.Args[k] = v
		}
		for k, v := range img.Args {
			named_img.Args[k] = v
		}

		imgs[i] = named_img
	}

	return imgs, nil
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

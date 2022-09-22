package clade

import (
	"errors"
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/distribution/distribution/reference"
	"golang.org/x/exp/slices"
)

type Image struct {
	Tags []string
	From string
	Args map[string]string

	Dockerfile  string
	ContextPath string
}

type Port struct {
	Name string
	Args map[string]string

	Dockerfile  string
	ContextPath string

	Images []Image
}

type NamedImage struct {
	reference.Named
	Tags []string
	From reference.NamedTagged
	Args map[string]string

	Dockerfile  string
	ContextPath string
}

func (i *NamedImage) Tagged() (reference.NamedTagged, error) {
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

func (p *Port) ParseImages() ([]*NamedImage, error) {
	name, err := reference.ParseNamed(p.Name)
	if err != nil {
		return nil, err
	}

	imgs := make([]*NamedImage, len(p.Images))
	for i, img := range p.Images {
		from, err := ParseReference(img.From)
		if err != nil {
			return nil, err
		}

		from_tagged, ok := from.(reference.NamedTagged)
		if !ok {
			return nil, errors.New("reference for \"from\" field must be tagged")
		}

		named_img := &NamedImage{
			Named: name,
			Tags:  slices.Clone(img.Tags),
			From:  from_tagged,
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

func ParseSemver(s string) (semver.Version, error) {
	ss := strings.SplitN(s, "-", 2)
	v, err := semver.ParseTolerant(ss[0])
	if err != nil {
		return semver.Version{}, err
	}

	if len(ss) > 1 {
		v.Pre = append(v.Pre, semver.PRVersion{VersionStr: ss[1]})
	}

	return v, nil
}

func DeduplicateBySemver(lhs *[]string, rhs *[]string) error {
	highest := func(vs []string) (semver.Version, error) {
		rst := semver.Version{}
		for _, v := range vs {
			sv, err := ParseSemver(v)
			if err != nil {
				return rst, fmt.Errorf("failed to parse semver: %w", err)
			}

			if rst.LT(sv) {
				rst = sv
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

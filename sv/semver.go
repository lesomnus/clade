package sv

import (
	"strings"

	"github.com/blang/semver/v4"
)

type Version struct {
	semver.Version
	Source string
}

func (s *Version) String() string {
	if s.Source != "" {
		return s.Source
	}

	return s.Version.String()
}

func Parse(s string) (Version, error) {
	ss := strings.SplitN(s, "-", 2)
	v, err := semver.ParseTolerant(ss[0])
	if err != nil {
		return Version{}, err
	}

	if len(ss) > 1 {
		v.Pre = append(v.Pre, semver.PRVersion{VersionStr: ss[1]})
	}

	return Version{Version: v, Source: s}, nil
}

func ParseStrict(s string) (Version, error) {
	v, err := semver.Parse(s)
	if err != nil {
		return Version{}, err
	}

	return Version{Version: v, Source: s}, nil
}

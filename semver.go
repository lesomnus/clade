package clade

import (
	"github.com/blang/semver/v4"
)

func ToSemver(v string, vs ...string) []semver.Version {
	rst := make([]semver.Version, 0, len(vs)+1)

	for _, c := range append([]string{v}, vs...) {
		sv, err := semver.ParseTolerant(c)
		if err != nil {
			continue
		}

		rst = append(rst, sv)
	}

	return rst
}

func SemverLatest(v semver.Version, vs ...semver.Version) semver.Version {
	for _, c := range vs {
		if v.LT(c) {
			v = c
		}
	}

	return v
}

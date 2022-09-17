package clade

import (
	"github.com/blang/semver/v4"
	"golang.org/x/exp/slices"
)

func ToSemver(v string, vs ...string) []semver.Version {
	rst := make([]semver.Version, 0, len(vs))

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

func SemverMajorN(n int, vs ...semver.Version) []semver.Version {
	majors := make([]uint64, len(vs))
	for _, c := range vs {
		majors = append(majors, c.Major)
	}

	majors = slices.Compact(majors)
	slices.Sort(majors)

	if len(majors) > n {
		majors = majors[len(majors)-n:]
	}

	rst := make([]semver.Version, 0, len(vs))
	for _, v := range vs {
		if slices.Contains(majors, v.Major) {
			rst = append(rst, v)
		}
	}

	return rst
}

func SemverMinorN(n int, vs ...semver.Version) []semver.Version {
	majors := make(map[uint64][]uint64)
	for _, c := range vs {
		minors, ok := majors[c.Major]
		if !ok {
			minors = make([]uint64, 0)
		}

		majors[c.Major] = append(minors, c.Minor)
	}

	rst := make([]semver.Version, 0, len(vs))
	for major, minors := range majors {
		minors := slices.Compact(minors)
		slices.Sort(minors)

		if len(minors) > n {
			minors = minors[len(minors)-n:]
		}

		for _, v := range vs {
			if v.Major != major {
				continue
			}

			if slices.Contains(minors, v.Minor) {
				rst = append(rst, v)
			}
		}
	}

	return rst
}

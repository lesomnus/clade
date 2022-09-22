package plf

import (
	"github.com/blang/semver/v4"
	"golang.org/x/exp/slices"
)

func ToSemver(vs ...string) []semver.Version {
	rst := make([]semver.Version, 0, len(vs))

	for _, v := range vs {
		sv, err := semver.ParseTolerant(v)
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
	if n == 0 {
		return vs
	}

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
	if n == 0 {
		return vs
	}

	group := make(map[uint64][]uint64)
	for _, c := range vs {
		minors, ok := group[c.Major]
		if !ok {
			minors = make([]uint64, 0)
		}

		group[c.Major] = append(minors, c.Minor)
	}

	rst := make([]semver.Version, 0, len(vs))
	for major, minors := range group {
		minors = slices.Compact(minors)
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

func SemverPatchN(n int, vs ...semver.Version) []semver.Version {
	if n == 0 {
		return vs
	}

	group := make(map[uint64]map[uint64][]uint64)
	for _, c := range vs {
		minors, ok := group[c.Major]
		if !ok {
			minors = make(map[uint64][]uint64)
			group[c.Major] = minors
		}

		patches, ok := minors[c.Minor]
		if !ok {
			patches = make([]uint64, 0)
		}

		minors[c.Minor] = append(patches, c.Patch)
	}

	rst := make([]semver.Version, 0, len(vs))
	for major, minors := range group {
		for minor, patches := range minors {
			patches = slices.Compact(patches)
			slices.Sort(patches)

			if len(patches) > n {
				patches = patches[len(patches)-n:]
			}

			for _, v := range vs {
				if v.Major != major {
					continue
				}
				if v.Minor != minor {
					continue
				}

				if slices.Contains(patches, v.Patch) {
					rst = append(rst, v)
				}
			}
		}
	}

	return rst
}

func SemverN(major int, minor int, patch int, vs ...semver.Version) []semver.Version {
	vs = SemverMajorN(major, vs...)
	vs = SemverMinorN(minor, vs...)
	vs = SemverPatchN(patch, vs...)

	return vs
}

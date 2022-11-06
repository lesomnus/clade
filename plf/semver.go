package plf

import (
	"github.com/lesomnus/clade/sv"
	"golang.org/x/exp/slices"
)

func noNil[T any](vs []*T) []*T {
	cursor := 0
	for _, v := range vs {
		if v == nil {
			continue
		}

		vs[cursor] = v
		cursor++
	}

	return vs[:cursor]
}

func Semver(ss ...string) []*sv.Version {
	rst := make([]*sv.Version, 0, len(ss))

	for _, s := range ss {
		v, err := sv.Parse(s)
		if err != nil {
			continue
		}

		rst = append(rst, &v)
	}

	return rst
}

func SemverFinalized(vs ...*sv.Version) []*sv.Version {
	vs = noNil(vs)

	for i, v := range vs {
		if len(v.Build) == 0 && len(v.Pre) == 0 {
			continue
		}

		vs[i] = nil
	}

	return noNil(vs)
}

func SemverLatest(v *sv.Version, vs ...*sv.Version) *sv.Version {
	for _, c := range vs {
		if v.LT(c.Version) {
			v = c
		}
	}

	return v
}

func SemverMajorN(n int, vs ...*sv.Version) []*sv.Version {
	vs = noNil(vs)
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

	rst := make([]*sv.Version, 0, len(vs))
	for _, v := range vs {
		if slices.Contains(majors, v.Major) {
			rst = append(rst, v)
		}
	}

	return rst
}

func SemverMinorN(n int, vs ...*sv.Version) []*sv.Version {
	vs = noNil(vs)
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

	rst := make([]*sv.Version, 0, len(vs))
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

func SemverPatchN(n int, vs ...*sv.Version) []*sv.Version {
	vs = noNil(vs)
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

	rst := make([]*sv.Version, 0, len(vs))
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

func SemverN(major int, minor int, patch int, vs ...*sv.Version) []*sv.Version {
	vs = noNil(vs)
	vs = SemverMajorN(major, vs...)
	vs = SemverMinorN(minor, vs...)
	vs = SemverPatchN(patch, vs...)

	return vs
}

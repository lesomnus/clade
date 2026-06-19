package tag

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/goccy/go-yaml"
)

func init() {
	Register("semver", newSemver)
}

// semverConfig is the target spec for the semver strategy.
//
//	target:
//	  kind: semver
//	  last-major: 2     # keep the latest 2 major lines (0 = all)
//	  last-minor: 3     # keep the latest 3 minor lines per major (0 = all)
//	  pre-release: alpine # only tags with this exact semver pre-release
type semverConfig struct {
	LastMajor  int    `yaml:"last-major"`
	LastMinor  int    `yaml:"last-minor"`
	PreRelease string `yaml:"pre-release"`
}

type semverSelector struct {
	lastMajor  int
	lastMinor  int
	preRelease string
}

func newSemver(params []byte) (Selector, error) {
	var cfg semverConfig
	if err := yaml.Unmarshal(params, &cfg); err != nil {
		return nil, fmt.Errorf("decode semver target: %w", err)
	}

	return &semverSelector{
		lastMajor:  cfg.LastMajor,
		lastMinor:  cfg.LastMinor,
		preRelease: cfg.PreRelease,
	}, nil
}

type semverTag struct {
	tag     string
	version *semver.Version
}

// Select keeps, for the latest lastMajor major lines, the latest lastMinor
// minor lines, each represented by its newest patch.
func (s *semverSelector) Select(tags []string) ([]Matched, error) {
	// Collapse to the newest version per (major, minor) line.
	lines := map[[2]uint64]semverTag{}
	for _, t := range tags {
		v, err := semver.NewVersion(t)
		if err != nil {
			continue // ignore tags that are not semver
		}
		if v.Prerelease() != s.preRelease {
			continue // keep only the exact pre-release ("" = none)
		}

		key := [2]uint64{v.Major(), v.Minor()}
		if cur, ok := lines[key]; !ok || v.GreaterThan(cur.version) {
			lines[key] = semverTag{tag: t, version: v}
		}
	}

	// Group lines by major.
	by_major := map[uint64][]semverTag{}
	for _, line := range lines {
		major := line.version.Major()
		by_major[major] = append(by_major[major], line)
	}

	majors := make([]uint64, 0, len(by_major))
	for major := range by_major {
		majors = append(majors, major)
	}
	sort.Slice(majors, func(i, j int) bool { return majors[i] > majors[j] })
	if s.lastMajor > 0 && len(majors) > s.lastMajor {
		majors = majors[:s.lastMajor]
	}

	out := []Matched{}
	for _, major := range majors {
		group := by_major[major]
		sort.Slice(group, func(i, j int) bool { return group[i].version.GreaterThan(group[j].version) })
		if s.lastMinor > 0 && len(group) > s.lastMinor {
			group = group[:s.lastMinor]
		}
		for _, line := range group {
			out = append(out, Matched{Tag: line.tag, Data: line.version})
		}
	}
	return out, nil
}

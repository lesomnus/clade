package plf_test

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/sv"
	"github.com/stretchr/testify/require"
)

func TestSemverMajorN(t *testing.T) {
	type Input struct {
		n  int
		vs []*sv.Version
	}

	require := require.New(t)

	tcs := []struct {
		input    Input
		expected []*sv.Version
	}{
		{
			input: Input{
				n: 1,
				vs: []*sv.Version{
					{Version: semver.Version{Major: 0, Minor: 1, Patch: 0}, Source: ""},
					{Version: semver.Version{Major: 2, Minor: 3, Patch: 4}, Source: ""},
					{Version: semver.Version{Major: 2, Minor: 3, Patch: 3}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 0, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 0, Minor: 2, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 1, Patch: 1}, Source: ""},
				},
			},
			expected: []*sv.Version{
				{Version: semver.Version{Major: 2, Minor: 3, Patch: 3}, Source: ""},
				{Version: semver.Version{Major: 2, Minor: 3, Patch: 4}, Source: ""},
			},
		},
		{
			input: Input{
				n: 2,
				vs: []*sv.Version{
					{Version: semver.Version{Major: 0, Minor: 1, Patch: 0}, Source: ""},
					{Version: semver.Version{Major: 2, Minor: 3, Patch: 4}, Source: ""},
					{Version: semver.Version{Major: 2, Minor: 3, Patch: 3}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 0, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 0, Minor: 2, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 1, Patch: 1}, Source: ""},
				},
			},
			expected: []*sv.Version{
				{Version: semver.Version{Major: 1, Minor: 0, Patch: 1}, Source: ""},
				{Version: semver.Version{Major: 1, Minor: 1, Patch: 1}, Source: ""},
				{Version: semver.Version{Major: 2, Minor: 3, Patch: 3}, Source: ""},
				{Version: semver.Version{Major: 2, Minor: 3, Patch: 4}, Source: ""},
			},
		},
	}
	for _, tc := range tcs {
		actual := plf.SemverMajorN(tc.input.n, tc.input.vs...)
		require.ElementsMatch(tc.expected, actual)
	}
}

func TestSemverMinorN(t *testing.T) {
	type Input struct {
		n  int
		vs []*sv.Version
	}

	require := require.New(t)

	tcs := []struct {
		input    Input
		expected []*sv.Version
	}{
		{
			input: Input{
				n: 2,
				vs: []*sv.Version{
					{Version: semver.Version{Major: 0, Minor: 1, Patch: 0}, Source: ""},
					{Version: semver.Version{Major: 2, Minor: 3, Patch: 4}, Source: ""},
					{Version: semver.Version{Major: 2, Minor: 3, Patch: 3}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 0, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 0, Patch: 5}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 2, Patch: 5}, Source: ""},
					{Version: semver.Version{Major: 0, Minor: 1, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 1, Patch: 1}, Source: ""},
					{Version: semver.Version{Major: 1, Minor: 1, Patch: 0}, Source: ""},
				},
			},
			expected: []*sv.Version{
				{Version: semver.Version{Major: 0, Minor: 1, Patch: 0}, Source: ""},
				{Version: semver.Version{Major: 0, Minor: 1, Patch: 1}, Source: ""},
				{Version: semver.Version{Major: 1, Minor: 1, Patch: 0}, Source: ""},
				{Version: semver.Version{Major: 1, Minor: 1, Patch: 1}, Source: ""},
				{Version: semver.Version{Major: 1, Minor: 2, Patch: 5}, Source: ""},
				{Version: semver.Version{Major: 2, Minor: 3, Patch: 3}, Source: ""},
				{Version: semver.Version{Major: 2, Minor: 3, Patch: 4}, Source: ""},
			},
		},
	}
	for _, tc := range tcs {
		actual := plf.SemverMinorN(tc.input.n, tc.input.vs...)
		require.ElementsMatch(tc.expected, actual)
	}
}

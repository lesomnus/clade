package clade_test

import (
	"testing"

	"github.com/blang/semver/v4"
	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/pipeline"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/sv"
	"github.com/stretchr/testify/require"
)

func TestParseReference(t *testing.T) {
	tcs := []struct {
		ref      string
		expected string
	}{
		{
			ref:      "cr.io/repo/clade:tag",
			expected: "tag",
		},
		{
			ref:      "cr.io/repo/clade:(tag)",
			expected: "(tag)",
		},
	}
	for _, tc := range tcs {
		t.Run("::"+tc.ref, func(t *testing.T) {
			require := require.New(t)

			named, err := clade.ParseReference(tc.ref)
			require.NoError(err)

			tagged, ok := named.(reference.NamedTagged)
			require.True(ok)

			require.Equal(tc.expected, tagged.Tag())
		})
	}
}

func TestRefNamedPipelineTagged(t *testing.T) {
	require := require.New(t)

	named, err := clade.ParseReference("cr.io/repo/clade:( localTags | toSemver | semverLatest )")
	require.NoError(err)

	tagged, ok := named.(clade.RefNamedPipelineTagged)
	require.True(ok)

	exe := pipeline.Executor{
		Funcs: pipeline.FuncMap{
			"localTags":    func() []string { return []string{"1.0", "1.1", "2.0"} },
			"toSemver":     plf.ToSemver,
			"semverLatest": plf.SemverLatest,
		},
	}

	v, err := exe.Execute(tagged.Pipeline())
	require.NoError(err)
	require.Len(v, 1)
	require.Equal(&sv.Version{Version: semver.Version{Major: 2}, Source: "2.0"}, v[0])
}

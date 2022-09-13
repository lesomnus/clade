package clade_test

import (
	"strings"
	"testing"
	"text/template"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
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
			ref:      "cr.io/repo/clade:/tag/",
			expected: "/tag/",
		},
		{
			ref:      "cr.io/repo/clade:{tag}",
			expected: "{tag}",
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

	named, err := clade.ParseReference("cr.io/repo/clade:{ localTags | semverLatest }")
	require.NoError(err)

	tagged, ok := named.(clade.RefNamedPipelineTagged)
	require.True(ok)

	tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"localTags":    func() []string { return []string{"1.0", "1.1", "2.0"} },
			"semverLatest": clade.SemverStringLatest,
		}).
		Parse(tagged.PipelineExpr())
	require.NoError(err)

	sb := strings.Builder{}
	err = tmpl.Execute(&sb, nil)
	require.NoError(err)
	require.Equal("2.0.0", sb.String())
}

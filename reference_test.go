package clade_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseRefNamedTagged(t *testing.T) {
	t.Run("tagged", func(t *testing.T) {
		require := require.New(t)

		ref, err := clade.ParseRefNamedTagged("cr.io/foo/bar:baz")
		require.NoError(err)
		require.Equal("cr.io/foo/bar", ref.Name())
		require.Equal("baz", ref.Tag())
		require.Equal("cr.io/foo/bar:baz", ref.String())
	})

	t.Run("pipeline tagged", func(t *testing.T) {
		require := require.New(t)

		ref, err := clade.ParseRefNamedTagged(`cr.io/foo/bar:(pass "baz")`)
		require.NoError(err)

		pl_ref, ok := ref.(clade.RefNamedPipelineTagged)
		require.True(ok)
		require.Equal("cr.io/foo/bar", pl_ref.Name())
		require.Equal(`(pass "baz")`, pl_ref.Tag())
		require.Equal(`cr.io/foo/bar:(pass "baz")`, pl_ref.String())

		require.Equal("cr.io", reference.Domain(pl_ref))
		require.Equal("foo/bar", reference.Path(pl_ref))
	})

	t.Run("domain with port", func(t *testing.T) {
		require := require.New(t)

		ref, err := clade.ParseRefNamedTagged(`cr.io:42/foo/bar:baz`)
		require.NoError(err)
		require.Equal("cr.io:42/foo/bar", ref.Name())
		require.Equal("baz", ref.Tag())
		require.Equal("cr.io:42/foo/bar:baz", ref.String())
	})

	t.Run("many colons", func(t *testing.T) {
		require := require.New(t)

		ref, err := clade.ParseRefNamedTagged(`cr.io:42/foo/bar:(tagsOf "cr.io:42/baz")`)
		require.NoError(err)

		pl_ref, ok := ref.(clade.RefNamedPipelineTagged)
		require.True(ok)
		require.Equal("cr.io:42/foo/bar", pl_ref.Name())
		require.Equal(`(tagsOf "cr.io:42/baz")`, pl_ref.Tag())
		require.Equal(`cr.io:42/foo/bar:(tagsOf "cr.io:42/baz")`, pl_ref.String())

		require.Equal("cr.io:42", reference.Domain(pl_ref))
		require.Equal("foo/bar", reference.Path(pl_ref))
	})

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc  string
			input string
			msg   string
		}{
			{
				desc:  "no tag",
				input: "cr.io/foo/bar",
				msg:   "no tag",
			},
			{
				desc:  "no tag with colon",
				input: "cr.io/foo/bar:",
				msg:   "no tag",
			},
			{
				desc:  "invalid name",
				input: "cr.io/foo/hey ho:baz",
				msg:   "invalid",
			},
			{
				desc:  "invalid tag",
				input: "cr.io/foo/bar:hey ho",
				msg:   "invalid",
			},
			{
				desc:  "invalid pipeline tag",
				input: "cr.io/foo/bar:(no string)",
				msg:   "invalid",
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				_, err := clade.ParseRefNamedTagged(tc.input)
				require.ErrorContains(err, tc.msg)
			})
		}
	})
}

func TestAsRefNamedPipelineTagged(t *testing.T) {
	t.Run("returns same one for pipeline tagged", func(t *testing.T) {
		require := require.New(t)

		ref, err := clade.ParseRefNamedTagged(`cr.io/foo/bar:(baz "answer" 42)`)
		require.NoError(err)

		pl_ref := clade.AsRefNamedPipelineTagged(ref)
		require.Equal(`(baz "answer" 42)`, pl_ref.Tag())
	})

	t.Run("tag is transformed into pipeline expression", func(t *testing.T) {
		require := require.New(t)

		ref, err := clade.ParseRefNamedTagged("cr.io/foo/bar:baz")
		require.NoError(err)

		pl_ref := clade.AsRefNamedPipelineTagged(ref)
		require.Equal("baz", pl_ref.Tag())
	})
}

func TestRefNamedPipelineTaggedUnmarshalYAML(t *testing.T) {
	tcs := []struct {
		desc     string
		expected string
		input    string
	}{
		{
			desc:     "single line representation with tag string",
			expected: "cr.io/foo/bar:baz",
			input:    "cr.io/foo/bar:baz",
		},
		{
			desc:     "single line representation with pipeline tagged",
			expected: `cr.io/foo/bar:(a "pi" 3.14)`,
			input:    `cr.io/foo/bar:(a "pi" 3.14)`,
		},
		{
			desc:     "object representation with with tag string",
			expected: "cr.io/foo/bar:baz",
			input: `
name: cr.io/foo/bar
tag: baz
`,
		},
		{
			desc:     "single line representation with pipeline tagged",
			expected: `cr.io/foo/bar:(a "pi" 3.14)`,
			input: `
name: cr.io/foo/bar
tag: (a "pi" 3.14)
`,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)
			expected := clade.AsRefNamedPipelineTagged(must(clade.ParseRefNamedTagged(tc.expected)))

			// This must be of type *clade.refNamedPipelineTagged.
			actual := clade.AsRefNamedPipelineTagged(must(clade.ParseRefNamedTagged("x.x/x/x:x")))
			err := yaml.Unmarshal([]byte(tc.input), actual)
			require.NoError(err)
			require.Equal(expected, actual)
		})
	}

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc  string
			msg   string
			input string
		}{
			{
				desc:  "invalid reference format",
				msg:   "must be canonical",
				input: "nodomain/foo/bar:baz",
			},
			{
				desc: "invalid type of scalar",
				msg:  "invalid",
				input: `!!binary |
  $
`,
			},
			{
				desc:  "invalid type",
				msg:   "invalid",
				input: "[cr.io/foo/bar:baz]",
			},
			{
				desc: "invalid map type",
				msg:  "cannot unmarshal",
				input: `
name: [cr.io/foo/bar]
tag: baz
`,
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				// This must be of type *clade.refNamedPipelineTagged.
				actual := clade.AsRefNamedPipelineTagged(must(clade.ParseRefNamedTagged("x.x/x/x:x")))
				err := yaml.Unmarshal([]byte(tc.input), actual)
				require.ErrorContains(err, tc.msg)
			})
		}
	})
}

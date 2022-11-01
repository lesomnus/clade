package clade_test

import (
	"testing"

	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

func TestPortUnmarshalYAML(t *testing.T) {
	t.Run("well formed", func(t *testing.T) {
		require := require.New(t)

		var port clade.Port
		err := yaml.Unmarshal([]byte(`
name: cr.io/foo/bar
args:
  USERNAME: hypnos 

dockerfile: ./path/to/dockerfile
context: ./../ctx

images:
  - tags: [a, b, c]
    from: hub.io/foo/bar:a
`), &port)

		require.NoError(err)
		require.Equal("cr.io/foo/bar", port.Name.String())

		require.Contains(port.Args, "USERNAME")
		require.Equal(port.Args["USERNAME"], "hypnos")

		require.Equal(port.Dockerfile, "./path/to/dockerfile")
		require.Equal(port.ContextPath, "./../ctx")

		require.Len(port.Images, 1)
		require.Equal("cr.io/foo/bar", port.Images[0].Name())
		require.Equal([]string{"a", "b", "c"}, port.Images[0].Tags)
		require.Equal("hub.io/foo/bar:a", port.Images[0].From.String())
		require.Empty(port.Images[0].Args)
		require.Empty(port.Images[0].Dockerfile)
		require.Empty(port.Images[0].ContextPath)
	})

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc  string
			input string
			msgs  []string
		}{
			{
				desc:  "invalid port format",
				input: "args: [SpongeBob SquarePants]",
				msgs:  []string{"seq", "map[string]string"},
			},
			{
				desc: "name not string",
				input: `name:
  - Somnus`,
				msgs: []string{"seq", "string"},
			},
			{
				desc:  "name is invalid reference format",
				input: "name: no/domain",
				msgs:  []string{"repo", "canonical"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var port clade.Port
				err := yaml.Unmarshal([]byte(tc.input), &port)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

func TestDeduplicateBySemver(t *testing.T) {
	type Pair struct {
		lhs []string
		rhs []string
	}

	tcs := []struct {
		given    Pair
		expected Pair
	}{
		{
			given: Pair{
				lhs: []string{"1.0.0", "1.0", "1"},
				rhs: []string{"1.0.0", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.0.0", "1.0", "1"},
				rhs: []string{},
			},
		},
		{
			given: Pair{
				lhs: []string{"1.0.1", "1.0", "1"},
				rhs: []string{"1.0.0", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.0.1", "1.0", "1"},
				rhs: []string{"1.0.0"},
			},
		},
		{
			given: Pair{
				lhs: []string{"1.0.0", "1.0", "1"},
				rhs: []string{"1.0.1", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.0.0"},
				rhs: []string{"1.0.1", "1.0", "1"},
			},
		},
		{
			given: Pair{
				lhs: []string{"1.1.1", "1.1", "1"},
				rhs: []string{"1.0.2", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.1.1", "1.1", "1"},
				rhs: []string{"1.0.2", "1.0"},
			},
		},
	}
	for _, tc := range tcs {
		actual := Pair{
			lhs: slices.Clone(tc.given.lhs),
			rhs: slices.Clone(tc.given.rhs),
		}

		err := clade.DeduplicateBySemver(&actual.lhs, &actual.rhs)
		require.NoError(t, err)
		require.ElementsMatch(t, tc.expected.lhs, actual.lhs)
		require.ElementsMatch(t, tc.expected.rhs, actual.rhs)
	}
}

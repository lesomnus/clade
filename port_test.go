package clade_test

import (
	"os"
	"path/filepath"
	"testing"

	ba "github.com/lesomnus/boolal"
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
platform: x & y | z

images:
  - tags: [a, b, c]
    from: hub.io/foo/bar:a

  - platform: a | b & c
`), &port)

		require.NoError(err)
		require.Equal("cr.io/foo/bar", port.Name.String())

		require.Contains(port.Args, "USERNAME")
		require.Equal("hypnos", port.Args["USERNAME"])

		require.Equal("./path/to/dockerfile", port.Dockerfile)
		require.Equal("./../ctx", port.ContextPath)
		require.Equal(ba.And("x", "y").Or("z"), port.Platform)

		require.Len(port.Images, 2)
		require.Equal("cr.io/foo/bar", port.Images[0].Name())

		tags := make([]string, len(port.Images[0].Tags))
		for i, tag := range port.Images[0].Tags {
			tags[i] = tag.String()
		}
		require.ElementsMatch([]string{"a", "b", "c"}, tags)

		require.Equal("hub.io/foo/bar:a", port.Images[0].From.String())
		require.Contains(port.Images[0].Args, "USERNAME")
		require.Equal("hypnos", port.Images[0].Args["USERNAME"].String())
		require.Empty(port.Images[0].Dockerfile)
		require.Empty(port.Images[0].ContextPath)

		require.Equal(ba.And("x", "y").Or("z"), port.Images[0].Platform)
		require.Equal(ba.Or("a", "b").And("c"), port.Images[1].Platform)
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
				desc:  "name not string",
				input: "name:\n  - Somnus",
				msgs:  []string{"seq", "string"},
			},
			{
				desc:  "name is invalid reference format",
				input: "name: no/domain",
				msgs:  []string{"repo", "canonical"},
			},
			{
				desc:  "platform is invalid syntax",
				input: "name: cr.io/foo/bar\nplatform: x && y || z",
				msgs:  []string{"platform:"},
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

func TestReadPort(t *testing.T) {
	require := require.New(t)

	data := `
name: cr.io/foo/bar
args:
  YEAR: 2009

images:
  - tags: [a, b, c]
    from: hub.io/foo/bar:baz

  - tags: [inglourious]
    from: hub.io/foo/bar:basterds
    args:
      VILLAIN: Hans Landa
    dockerfile: 35mm Nitrate Film
    context: Le Gamaar cinema
`

	dir := os.TempDir()
	tmp, err := os.CreateTemp(dir, "")
	require.NoError(err)

	defer os.Remove(tmp.Name())

	_, err = tmp.Write([]byte(data))
	require.NoError(err)

	port, err := clade.ReadPort(tmp.Name())
	require.NoError(err)
	require.Len(port.Images, 2)
	require.Equal(port.Images[0].From.String(), "hub.io/foo/bar:baz")
	require.Equal(port.Images[1].From.String(), "hub.io/foo/bar:basterds")

	require.Equal(filepath.Join(dir, "Dockerfile"), port.Images[0].Dockerfile, "default dockerfile is {path}/Dockerfile")
	require.Equal(filepath.Join(dir, "."), port.Images[0].ContextPath, "default context is {path}")
	require.Contains(port.Images[0].Args, "YEAR", "root args are inherited")
	require.Equal("2009", port.Images[0].Args["YEAR"].String(), "root args are inherited")

	require.Equal(port.Images[1].Dockerfile, filepath.Join(dir, "35mm Nitrate Film"))
	require.Equal(port.Images[1].ContextPath, filepath.Join(dir, "Le Gamaar cinema"))
	require.Contains(port.Images[1].Args, "YEAR")
	require.Equal("2009", port.Images[1].Args["YEAR"].String())
	require.Contains(port.Images[1].Args, "VILLAIN")
	require.Equal("Hans Landa", port.Images[1].Args["VILLAIN"].String())
}

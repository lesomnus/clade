package clade_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/pl"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

func TestPortUnmarshalYAML(t *testing.T) {
	t.Run(".name", func(t *testing.T) {
		require := require.New(t)

		var port clade.Port
		err := yaml.Unmarshal([]byte("name: cr.io/repo/name"), &port)
		require.NoError(err)
		require.Equal("cr.io/repo/name", port.Name.String())
	})

	t.Run(".images", func(t *testing.T) {
		t.Run("can have empty value", func(t *testing.T) {
			require := require.New(t)

			var port clade.Port
			err := yaml.Unmarshal([]byte(`
name: cr.io/repo/name
images:
  - 
`), &port)
			require.NoError(err)
			require.Len(port.Images, 1)
			require.Nil(port.Images[0])
		})

		t.Run("`.skip` is true if `$.skip` is true", func(t *testing.T) {
			require := require.New(t)

			var port clade.Port
			err := yaml.Unmarshal([]byte(`
name: cr.io/repo/name
images:
  - skip: false
`), &port)
			require.NoError(err)
			require.Len(port.Images, 1)
			require.False(port.Images[0].Skip)

			err = yaml.Unmarshal([]byte(`
name: cr.io/repo/name
skip: true
images:
  - skip: false
`), &port)
			require.NoError(err)
			require.Len(port.Images, 1)
			require.True(port.Images[0].Skip)
		})

		t.Run("`.platform` will be set by `$.platform` if it is empty", func(t *testing.T) {
			require := require.New(t)

			var port clade.Port
			err := yaml.Unmarshal([]byte(`
name: cr.io/repo/name
images:
  - platform: foo
`), &port)
			require.NoError(err)
			require.Len(port.Images, 1)
			require.NotNil(port.Images[0].Platform)
			require.Equal("foo", port.Images[0].Platform.Lhs.Var)

			err = yaml.Unmarshal([]byte(`
name: cr.io/repo/name
platform: bar
images:
  - name: cr.io/repo/baz
`), &port)
			require.NoError(err)
			require.Len(port.Images, 1)
			require.Equal("bar", port.Images[0].Platform.Lhs.Var)
		})

		t.Run("`.args` is merged from `$.args`", func(t *testing.T) {
			require := require.New(t)

			var port clade.Port
			err := yaml.Unmarshal([]byte(`
name: cr.io/repo/name
args:
  keyA: valA
  keyB: valB
images:
  - args:
      keyB: valFoo
      keyC: valC
`), &port)
			require.NoError(err)
			require.Len(port.Images, 1)
			require.Len(port.Images[0].Args, 3)
			require.Contains(port.Images[0].Args, "keyA")
			require.Contains(port.Images[0].Args, "keyB")
			require.Contains(port.Images[0].Args, "keyC")
			require.Equal(*(*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "valA")))), port.Images[0].Args["keyA"])
			require.Equal(*(*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "valFoo")))), port.Images[0].Args["keyB"], "no overwrite")
			require.Equal(*(*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "valC")))), port.Images[0].Args["keyC"])

		})
	})

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc string
			data string
			msgs []string
		}{
			{
				desc: "not a map",
				data: "string",
				msgs: []string{"string"},
			},
			{
				desc: "name is invalid",
				data: `{"name": "not_canonical"}`,
				msgs: []string{"canonical"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var actual clade.Port
				err := yaml.Unmarshal([]byte(tc.data), &actual)
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
  YEAR: 1995

images:
  - tags: [a, b, c]
    from: hub.io/foo/bar:baz

  - tags: [inglourious]
    from: hub.io/foo/bar:basterds
    args:
      YEAR: 2009
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
	require.Equal("hub.io/foo/bar", port.Images[0].From.Primary.String())
	require.Equal((*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "baz")))), port.Images[0].From.Primary.Tag)
	require.Equal("hub.io/foo/bar", port.Images[1].From.Primary.String())
	require.Equal((*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "basterds")))), port.Images[1].From.Primary.Tag)

	require.Equal(filepath.Join(dir, "Dockerfile"), port.Images[0].Dockerfile, "default dockerfile is {path}/Dockerfile")
	require.Equal(filepath.Join(dir, "."), port.Images[0].ContextPath, "default context is {path}")
	require.Contains(port.Images[0].Args, "YEAR", "root args are inherited")
	require.Equal(*(*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "1995")))), port.Images[0].Args["YEAR"], "root args are inherited")

	require.Equal(port.Images[1].Dockerfile, filepath.Join(dir, "35mm Nitrate Film"))
	require.Equal(port.Images[1].ContextPath, filepath.Join(dir, "Le Gamaar cinema"))
	require.Contains(port.Images[1].Args, "YEAR")
	require.Equal(*(*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "2009")))), port.Images[1].Args["YEAR"])
	require.Contains(port.Images[1].Args, "VILLAIN")
	require.Equal(*(*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "Hans Landa")))), port.Images[1].Args["VILLAIN"])
}

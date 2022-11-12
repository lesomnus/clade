package clade_test

import (
	"fmt"
	"testing"

	"github.com/distribution/distribution/reference"
	ba "github.com/lesomnus/boolal"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/pl"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func TestImageUnmarshalFromField(t *testing.T) {
	type TagExpr struct {
		name string
		tag  string
		pl   *pl.Pl
	}

	tcs := []struct {
		desc     string
		input    string
		expected TagExpr
	}{
		{
			desc:  "tagged",
			input: `from: "cr.io/repo/name:foo"`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "foo",
				pl:   pl.NewPl(must(pl.NewFn("pass", "foo"))),
			},
		},
		{
			desc:  "tagged with string pipeline expression",
			input: `from: cr.io/repo/name:( foo "bar" | baz )`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  `( foo "bar" | baz )`,
				pl: pl.NewPl(
					must(pl.NewFn("foo", "bar")),
					must(pl.NewFn("baz")),
				),
			},
		},
		{
			desc: "map",
			input: `from:
  name: cr.io/repo/name
  tag: foo
`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "foo",
				pl:   pl.NewPl(must(pl.NewFn("pass", "foo"))),
			},
		},
		{
			desc: "map and tag with string pipeline expression",
			input: `from:
  name: cr.io/repo/name
  tag: ( foo "bar" | baz )`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  `( foo "bar" | baz )`,
				pl: pl.NewPl(
					must(pl.NewFn("foo", "bar")),
					must(pl.NewFn("baz")),
				),
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			var img clade.Image
			err := yaml.Unmarshal([]byte(tc.input), &img)
			require.NoError(err)

			require.Equal(tc.expected.name, img.From.Name())
			require.Equal(tc.expected.tag, img.From.Tag())
			require.Equal(tc.expected.pl, img.From.Pipeline())
		})
	}

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc  string
			input string
			msgs  []string
		}{
			{
				desc:  "invalid image format",
				input: "tags: 42",
				msgs:  []string{"int", "42"},
			},
			{
				desc:  "invalid reference format for from",
				input: "from: cr.io/foo/bar",
				msgs:  []string{"no tag"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var img clade.Image
				err := yaml.Unmarshal([]byte(tc.input), &img)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

func TestImageUnmarshalPlatformField(t *testing.T) {
	tcs := []struct {
		desc     string
		input    string
		expected *ba.Expr
	}{
		{
			desc:     "nil if empty",
			input:    "",
			expected: nil,
		},
		{
			desc:     "with expr",
			input:    "x & y | z",
			expected: ba.And("x", "y").Or("z"),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			var img clade.Image
			err := yaml.Unmarshal([]byte(fmt.Sprintf("platform: %s", tc.input)), &img)
			require.NoError(err)
			require.Equal(tc.expected, img.Platform)
		})
	}

	t.Run("it fails if expression is invalid", func(t *testing.T) {
		require := require.New(t)

		var img clade.Image
		err := yaml.Unmarshal([]byte(`platform: x && y || z`), &img)
		require.ErrorContains(err, "platform:")
	})
}

func TestResolvedImageTagged(t *testing.T) {
	t.Run("tagged with first element", func(t *testing.T) {
		require := require.New(t)

		img := clade.ResolvedImage{
			Named: must(reference.ParseNamed("cr.io/foo/bar")),
			Tags:  []string{"a", "b", "c"},
		}

		tagged, err := img.Tagged()
		require.NoError(err)
		require.Equal("cr.io/foo/bar:a", tagged.String())
	})

	t.Run("fails if tag format invalid", func(t *testing.T) {
		require := require.New(t)

		img := clade.ResolvedImage{
			Named: must(reference.ParseNamed("cr.io/foo/bar")),
			Tags:  []string{"Edgar Wright"},
		}

		_, err := img.Tagged()
		require.ErrorContains(err, "invalid")
	})

	t.Run("fails if not tagged", func(t *testing.T) {
		require := require.New(t)

		img := clade.ResolvedImage{
			Named: must(reference.ParseNamed("cr.io/foo/bar")),
			Tags:  []string{},
		}

		_, err := img.Tagged()
		require.ErrorContains(err, "not tagged")
	})
}

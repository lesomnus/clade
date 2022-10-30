package clade_test

import (
	"testing"

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
			input: `from: "cr.io/repo/name:( foo bar | baz )"`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "( foo bar | baz )",
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
  tag: "( foo bar | baz )"`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "( foo bar | baz )",
				pl: pl.NewPl(
					must(pl.NewFn("foo", "bar")),
					must(pl.NewFn("baz")),
				),
			},
		},
		// 		{
		// 			desc: "map and tag with yaml pipeline expression",
		// 			input: `from:
		//   name: cr.io/repo/name
		//   tag:
		//     - foo bar
		//     - baz`,
		// 			expected: TagExpr{
		// 				name: "cr.io/repo/name",
		// 				tag:  "",
		// 				pl: pipeline.Pipeline{
		// 					&pipeline.Fn{Name: "foo", Args: []any{"bar"}},
		// 					&pipeline.Fn{Name: "baz", Args: []any{}},
		// 				},
		// 			},
		// 		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			var img clade.Image
			err := yaml.Unmarshal([]byte(tc.input), &img)
			require.NoError(err)

			require.Equal(tc.expected.name, img.From.Name())
			require.Equal(tc.expected.tag, img.From.Tag())
			require.ElementsMatch(tc.expected.pl, img.From.Pipeline())
		})
	}
}

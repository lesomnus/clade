package clade_test

import (
	"testing"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImageUnmarshalFromField(t *testing.T) {
	type TagExpr struct {
		name string
		tag  string
		pl   pipeline.Pipeline
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
				pl:   pipeline.Pipeline{&pipeline.Fn{Name: ">", Args: []any{"foo"}}},
			},
		},
		{
			desc:  "tagged with string pipeline expression",
			input: `from: "cr.io/repo/name:{ foo bar | baz }"`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "{ foo bar | baz }",
				pl: pipeline.Pipeline{
					&pipeline.Fn{Name: "foo", Args: []any{"bar"}},
					&pipeline.Fn{Name: "baz", Args: []any{}},
				},
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
				pl:   pipeline.Pipeline{&pipeline.Fn{Name: ">", Args: []any{"foo"}}},
			},
		},
		{
			desc: "map and tag with string pipeline expression",
			input: `from:
  name: cr.io/repo/name
  tag: "{ foo bar | baz }"`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "{ foo bar | baz }",
				pl: pipeline.Pipeline{
					&pipeline.Fn{Name: "foo", Args: []any{"bar"}},
					&pipeline.Fn{Name: "baz", Args: []any{}},
				},
			},
		},
		{
			desc: "map and tag with yaml pipeline expression",
			input: `from:
  name: cr.io/repo/name
  tag:
    - foo bar
    - baz`,
			expected: TagExpr{
				name: "cr.io/repo/name",
				tag:  "",
				pl: pipeline.Pipeline{
					&pipeline.Fn{Name: "foo", Args: []any{"bar"}},
					&pipeline.Fn{Name: "baz", Args: []any{}},
				},
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
			require.ElementsMatch(tc.expected.pl, img.From.Pipeline())
		})
	}
}

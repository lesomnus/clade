package pipeline_test

import (
	"testing"

	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalYaml(t *testing.T) {
	tcs := []struct {
		desc     string
		input    string
		expected pipeline.Pipeline
	}{
		{
			desc:  "expression for each sequence item",
			input: `["a b", "c d", "foo bar | baz", "e"]`,
			expected: pipeline.Pipeline{
				{Name: "a", Args: []any{"b"}},
				{Name: "c", Args: []any{"d"}},
				{Name: ">", Args: []any{pipeline.Pipeline{
					{Name: "foo", Args: []any{"bar"}},
					{Name: "baz", Args: []any{}},
				}}},
				{Name: "e", Args: []any{}},
			},
		},
		{
			desc:  "token sequence for each sequence item",
			input: `[["a", "b"], ["c", "d"], ["e"]]`,
			expected: pipeline.Pipeline{
				{Name: "a", Args: []any{"b"}},
				{Name: "c", Args: []any{"d"}},
				{Name: "e", Args: []any{}},
			},
		},
		{
			desc:  "mixed",
			input: `[["a", "b"], ["c", "d"], "foo bar | baz", ["e"]]`,
			expected: pipeline.Pipeline{
				{Name: "a", Args: []any{"b"}},
				{Name: "c", Args: []any{"d"}},
				{Name: ">", Args: []any{pipeline.Pipeline{
					{Name: "foo", Args: []any{"bar"}},
					{Name: "baz", Args: []any{}},
				}}},
				{Name: "e", Args: []any{}},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			pl := make([]*pipeline.Fn, 0)
			err := yaml.Unmarshal([]byte(tc.input), &pl)
			require.NoError(err)

			samePipeline(require, tc.expected, pl)
		})
	}
}

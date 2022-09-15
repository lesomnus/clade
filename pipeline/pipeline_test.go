package pipeline_test

import (
	"testing"

	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	exe := pipeline.Executor{
		Funcs: pipeline.FuncMap{
			"add": func(lhs int, rhs int) int {
				return lhs + rhs
			},
			"mul": func(lhs int, rhs int) int {
				return lhs * rhs
			},
		},
	}

	tcs := []struct {
		desc     string
		pl       pipeline.Pipeline
		expected any
	}{
		{
			desc: "pipe",
			pl: pipeline.Pipeline{
				{Name: "add", Args: []any{1, 2}},
				{Name: "add", Args: []any{3}},
			},
			expected: 6,
		},
		{
			desc: "implicit conversion",
			pl: pipeline.Pipeline{
				{Name: "add", Args: []any{"1", "2"}},
				{Name: "add", Args: []any{"3"}},
			},
			expected: 6,
		},
		{
			desc: "nested",
			pl: pipeline.Pipeline{
				{Name: "mul", Args: []any{1, pipeline.Pipeline{
					{Name: "add", Args: []any{2, 3}},
				}}},
				{Name: "add", Args: []any{4}},
			},
			expected: 9,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			actual, err := exe.Execute(tc.pl)
			require.NoError(err)
			require.Equal(tc.expected, actual)
		})
	}
}

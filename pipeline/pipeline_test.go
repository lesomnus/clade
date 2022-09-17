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
			"max": func(v int, vs ...int) int {
				for _, c := range vs {
					if v < c {
						v = c
					}
				}

				return v
			},
			"gen": func(end int) []int {
				rst := make([]int, end)
				for i := range rst {
					rst[i] = i
				}

				return rst
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
			desc: "pipe multiple values",
			pl: pipeline.Pipeline{
				{Name: "gen", Args: []any{5}},
				{Name: "max", Args: []any{2}},
			},
			expected: 4,
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
			desc: "variadic",
			pl: pipeline.Pipeline{
				{Name: "max", Args: []any{4, 2, 5, 8, 6}},
				{Name: "add", Args: []any{3}},
			},
			expected: 11,
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
		{
			desc: "nested with variadic",
			pl: pipeline.Pipeline{
				{Name: "mul", Args: []any{1, pipeline.Pipeline{
					{Name: "max", Args: []any{6, 4, 5}},
					{Name: "add", Args: []any{2}},
				}}},
				{Name: "add", Args: []any{4}},
			},
			expected: 12,
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

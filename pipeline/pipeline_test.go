package pipeline_test

import (
	"fmt"
	"testing"

	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
)

type StringSet map[string]struct{}

func (sl StringSet) String() string {
	return fmt.Sprint(map[string]struct{}(sl))
}

func TestExecute(t *testing.T) {
	exe := pipeline.Executor{
		Funcs: pipeline.FuncMap{
			"add": func(lhs int, rhs int) int {
				return lhs + rhs
			},
			"sum": func(is ...int) int {
				sum := 0
				for _, i := range is {
					sum += i
				}

				return sum
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
			"strings": func(ss ...string) StringSet {
				rst := make(StringSet)
				for _, s := range ss {
					rst[s] = struct{}{}
				}

				return rst
			},
			"concat": func(ss ...string) string {
				rst := ""
				for _, s := range ss {
					rst += s
				}

				return rst
			},
		},
	}

	tcs := []struct {
		desc     string
		pl       pipeline.Pipeline
		expected []any
	}{
		{
			desc: "built in: pass",
			pl: pipeline.Pipeline{
				{Name: ">", Args: []any{1, 2}},
				{Name: "sum", Args: []any{3}},
			},
			expected: []any{6},
		},
		{
			desc: "pipe",
			pl: pipeline.Pipeline{
				{Name: "add", Args: []any{1, 2}},
				{Name: "add", Args: []any{3}},
			},
			expected: []any{6},
		},
		{
			desc: "pipe multiple values",
			pl: pipeline.Pipeline{
				{Name: "gen", Args: []any{5}},
				{Name: "max", Args: []any{2}},
			},
			expected: []any{4},
		},
		{
			desc: "implicit conversion",
			pl: pipeline.Pipeline{
				{Name: "add", Args: []any{"1", "2"}},
				{Name: "add", Args: []any{"3"}},
			},
			expected: []any{6},
		},
		{
			desc: "variadic",
			pl: pipeline.Pipeline{
				{Name: "max", Args: []any{4, 2, 5, 8, 6}},
				{Name: "add", Args: []any{3}},
			},
			expected: []any{11},
		},
		{
			desc: "nested",
			pl: pipeline.Pipeline{
				{Name: "mul", Args: []any{1, pipeline.Pipeline{
					{Name: "add", Args: []any{2, 3}},
				}}},
				{Name: "add", Args: []any{4}},
			},
			expected: []any{9},
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
			expected: []any{12},
		},
		{
			desc: "convert to string with String()",
			pl: pipeline.Pipeline{
				{Name: "strings", Args: []any{"foo", "bar"}},
				{Name: "concat", Args: []any{"hello ", "world"}},
			},
			expected: []any{"hello world" + fmt.Sprint(StringSet{"foo": struct{}{}, "bar": struct{}{}})},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			actual, err := exe.Execute(tc.pl)
			require.NoError(err)
			require.ElementsMatch(tc.expected, actual)
		})
	}
}

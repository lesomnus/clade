package pipeline_test

import (
	"errors"
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

			"discard": func(vs ...any) {},
			"return3": func(vs ...any) (string, int, error) {
				return "", 42, nil
			},
			"return-no-error": func(vs ...any) (string, int) {
				return "", 42
			},
			"as-error": func(msg string) (string, error) {
				return "", errors.New(msg)
			},
		},
	}

	{
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
				desc: "convert to string with String() of map",
				pl: pipeline.Pipeline{
					{Name: "strings", Args: []any{"foo", "bar"}},
					{Name: "concat", Args: []any{"hello ", "world"}},
				},
				expected: []any{"hello world" + fmt.Sprint(StringSet{"foo": struct{}{}, "bar": struct{}{}})},
			},
			{
				desc: "convert to string from int",
				pl: pipeline.Pipeline{
					{Name: "concat", Args: []any{4, 2}},
				},
				expected: []any{"42"},
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

	{
		tcs := []struct {
			desc string
			pl   pipeline.Pipeline
			msgs []string
		}{
			{
				desc: "function must be known",
				pl:   pipeline.Pipeline{{Name: "not-exists", Args: []any{"foo", "bar"}}},
				msgs: []string{"unknown", "function"},
			},
			{
				desc: "error from function is returned as error",
				pl:   pipeline.Pipeline{{Name: "as-error", Args: []any{"morty"}}},
				msgs: []string{"morty"},
			},
			{
				desc: "function have to return at least one value",
				pl:   pipeline.Pipeline{{Name: "discard", Args: []any{"foo", "bar"}}},
				msgs: []string{"function", "return", "one"},
			},
			{
				desc: "function can return up to three values",
				pl:   pipeline.Pipeline{{Name: "return3", Args: []any{"foo", "bar"}}},
				msgs: []string{"function", "return", "two"},
			},
			{
				desc: "second return value must be an error type.",
				pl:   pipeline.Pipeline{{Name: "return-no-error", Args: []any{"foo", "bar"}}},
				msgs: []string{"second", "type", "error"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				_, err := exe.Execute(tc.pl)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	}
}

func TestHasFunction(t *testing.T) {
	require := require.New(t)

	pl := pipeline.Pipeline{
		{Name: "a", Args: []any{"x", pipeline.Pipeline{
			{Name: "b", Args: []any{"y"}},
		}, "z"}},
	}

	require.True(pl.HasFunction("a"), "a")
	require.True(pl.HasFunction("b"), "b")
	require.False(pl.HasFunction("z"), "z")
}

func TestReturn(t *testing.T) {
	require := require.New(t)

	expected := "bender"
	exe := pipeline.Executor{Funcs: make(pipeline.FuncMap)}

	actual, err := exe.Execute(pipeline.Return(expected))
	require.NoError(err)
	require.Len(actual, 1)
	require.Equal(expected, actual[0])
}

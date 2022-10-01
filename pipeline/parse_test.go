package pipeline_test

import (
	"bytes"
	"testing"

	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
)

func samePipeline(require *require.Assertions, expected pipeline.Pipeline, actual pipeline.Pipeline) {
	require.Equal(len(expected), len(actual))
	for i, cmd_actual := range actual {
		cmd_expected := expected[i]
		require.Equal(cmd_expected.Name, cmd_actual.Name)
		require.Equal(len(cmd_expected.Args), len(cmd_actual.Args))

		for j, arg_actual := range cmd_actual.Args {
			arg_expected := cmd_expected.Args[j]
			if nested_expected, ok := arg_expected.(pipeline.Pipeline); ok {
				nested_actual, ok := arg_actual.(pipeline.Pipeline)
				require.True(ok)
				samePipeline(require, nested_expected, nested_actual)
			} else {
				require.Equal(arg_expected, arg_actual)
			}
		}
	}
}

func TestParse(t *testing.T) {
	{
		tcs := []struct {
			input    string
			expected pipeline.Pipeline
		}{
			{
				input: "foo",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{}},
				},
			},
			{
				input: "foo | bar",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{}},
					{Name: "bar", Args: []interface{}{}},
				},
			},
			{
				input: "foo a `b` c | bar 1 2 `3`",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{"a", "b", "c"}},
					{Name: "bar", Args: []interface{}{"1", "2", "3"}},
				},
			},
			{
				input: "foo a b (baz) | bar 1 2 3",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{"a", "b", pipeline.Pipeline{
						{Name: "baz", Args: []interface{}{}},
					}}},
					{Name: "bar", Args: []interface{}{"1", "2", "3"}},
				},
			},
			{
				input: "foo a b (baz x y z) | bar 1 2 3",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{"a", "b", pipeline.Pipeline{
						{Name: "baz", Args: []interface{}{"x", "y", "z"}},
					}}},
					{Name: "bar", Args: []interface{}{"1", "2", "3"}},
				},
			},
			{
				input: "foo a (bar 1 2 3 | baz x y z) c",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{
						"a",
						pipeline.Pipeline{
							{Name: "bar", Args: []interface{}{"1", "2", "3"}},
							{Name: "baz", Args: []interface{}{"x", "y", "z"}},
						},
						"c"},
					},
				},
			},
			{
				input: "foo a b (bar 1 2 (baz x y z))",
				expected: pipeline.Pipeline{
					{Name: "foo", Args: []interface{}{"a", "b", pipeline.Pipeline{
						{Name: "bar", Args: []interface{}{"1", "2", pipeline.Pipeline{
							{Name: "baz", Args: []interface{}{"x", "y", "z"}},
						}}},
					}}},
				},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.input, func(t *testing.T) {
				require := require.New(t)

				pl, err := pipeline.Parse(bytes.NewReader([]byte(tc.input)))
				require.NoError(err)
				samePipeline(require, tc.expected, pl)
			})
		}
	}

	{
		tcs := []struct {
			desc  string
			input string
			msgs  []string
		}{
			{
				desc:  "there must be at least one function name",
				input: " ",
				msgs:  []string{"expected a function name"},
			},
			{
				desc:  "there must be at least one function name before pipe",
				input: "| a",
				msgs:  []string{"expected a function name"},
			},
			{
				desc:  "there must be at least one function name after pipe",
				input: "a |",
				msgs:  []string{"expected a function name"},
			},
			{
				desc:  "string literal cannot be a function name",
				input: "a b c | `d` e f",
				msgs:  []string{"expected a function name"},
			},
			{
				desc:  "nested pipeline not allowed to function name",
				input: "(a b c) d e",
				msgs:  []string{"expected a function name"},
			},
			{
				desc:  "root scope cannot be closed",
				input: "a b c)",
				msgs:  []string{"end of scope"},
			},
			{
				desc:  "nested pipeline must be closed",
				input: "a (b c",
				msgs:  []string{"scope", "not closed"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				_, err := pipeline.Parse(bytes.NewReader([]byte(tc.input)))
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	}
}

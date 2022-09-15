package pipeline_test

import (
	"testing"

	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
)

func TestReadToken(t *testing.T) {
	type Token struct {
		pos   int
		value string
	}

	tcs := []struct {
		input  string
		tokens []Token
	}{
		{
			input: "foo",
			tokens: []Token{
				{0, "foo"},
			},
		},
		{
			input: "  foo",
			tokens: []Token{
				{2, "foo"},
			},
		},
		{
			input: "foo bar",
			tokens: []Token{
				{0, "foo"},
				{1, "bar"},
			},
		},
		{
			input: "foo  bar",
			tokens: []Token{
				{0, "foo"},
				{2, "bar"},
			},
		},
		{
			input: " foo  bar",
			tokens: []Token{
				{1, "foo"},
				{2, "bar"},
			},
		},
		{
			input: " foo  bar baz",
			tokens: []Token{
				{1, "foo"},
				{2, "bar"},
				{1, "baz"},
			},
		},
		{
			input: " foo | bar",
			tokens: []Token{
				{1, "foo"},
				{1, "|"},
				{1, "bar"},
			},
		},
		{
			input: "foo|bar",
			tokens: []Token{
				{0, "foo"},
				{0, "|"},
				{0, "bar"},
			},
		},
		{
			input: "||",
			tokens: []Token{
				{0, "|"},
				{0, "|"},
			},
		},
		{
			input: "( )",
			tokens: []Token{
				{0, "("},
				{1, ")"},
			},
		},
		{
			input: "| |",
			tokens: []Token{
				{0, "|"},
				{1, "|"},
			},
		},
		{
			input: "  | |",
			tokens: []Token{
				{2, "|"},
				{1, "|"},
			},
		},
		{
			input: "  |  (  ",
			tokens: []Token{
				{2, "|"},
				{2, "("},
			},
		},
		{
			input: "  )  |  ",
			tokens: []Token{
				{2, ")"},
				{2, "|"},
			},
		},
		{
			input: "\"foo\"",
			tokens: []Token{
				{0, "\"foo\""},
			},
		},
		{
			input: "\"foo bar\"",
			tokens: []Token{
				{0, "\"foo bar\""},
			},
		},
		{
			input: "\"foo\\\"bar\"",
			tokens: []Token{
				{0, "\"foo\\\"bar\""},
			},
		},
		{
			input: "\"foo\\\"|bar\"",
			tokens: []Token{
				{0, "\"foo\\\"|bar\""},
			},
		},
		{
			input: "  \"foo\\\"|bar\"",
			tokens: []Token{
				{2, "\"foo\\\"|bar\""},
			},
		},
		{
			input: "foo\"bar",
			tokens: []Token{
				{0, "foo\"bar"},
			},
		},
		{
			input: "foo\"bar\"",
			tokens: []Token{
				{0, "foo\"bar\""},
			},
		},
		{
			input: "foo\"bar\"baz",
			tokens: []Token{
				{0, "foo\"bar\"baz"},
			},
		},
		{
			input: " \"foo\\\"bar\" | baz",
			tokens: []Token{
				{1, "\"foo\\\"bar\""},
				{1, "|"},
				{1, "baz"},
			},
		},
		{
			input: "(foo\"bar\"|baz)",
			tokens: []Token{
				{0, "("},
				{0, "foo\"bar\""},
				{0, "|"},
				{0, "baz"},
				{0, ")"},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			require := require.New(t)

			rest := tc.input
			for _, token := range tc.tokens {
				pos, value, err := pipeline.ReadToken(rest)
				require.NoError(err)
				require.Equal(token, Token{pos, value})

				rest = rest[pos+len(value):]
			}
		})
	}
}

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
			input: "foo a b c | bar 1 2 3",
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

			cmds, err := pipeline.Parse(tc.input)
			require.NoError(err)
			samePipeline(require, tc.expected, cmds)
		})
	}
}

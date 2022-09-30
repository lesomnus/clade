package pipeline_test

import (
	"bytes"
	"testing"

	"github.com/lesomnus/clade/pipeline"
	"github.com/stretchr/testify/require"
)

func TestLex(t *testing.T) {
	tcs := []struct {
		input    string
		expected []pipeline.TokenItem
	}{
		{
			input: "()|",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenLeftParen, "("},
				{pipeline.Pos{1, 1}, pipeline.TokenRightParen, ")"},
				{pipeline.Pos{1, 2}, pipeline.TokenPipe, "|"},
			},
		},
		{
			input: "3.14",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenText, "3.14"},
			},
		},
		{
			input: "+3.14",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenText, "+3.14"},
			},
		},
		{
			input: "-3.14",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenText, "-3.14"},
			},
		},
		{
			input: "`()|`",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenString, "`()|`"},
			},
		},
		{
			input: "()|foo `bar` ",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenLeftParen, "("},
				{pipeline.Pos{1, 1}, pipeline.TokenRightParen, ")"},
				{pipeline.Pos{1, 2}, pipeline.TokenPipe, "|"},
				{pipeline.Pos{1, 3}, pipeline.TokenText, "foo"},
				{pipeline.Pos{1, 7}, pipeline.TokenString, "`bar`"},
			},
		},
		{
			input: "(foo|`bar`)",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenLeftParen, "("},
				{pipeline.Pos{1, 1}, pipeline.TokenText, "foo"},
				{pipeline.Pos{1, 4}, pipeline.TokenPipe, "|"},
				{pipeline.Pos{1, 5}, pipeline.TokenString, "`bar`"},
				{pipeline.Pos{1, 10}, pipeline.TokenRightParen, ")"},
			},
		},
		{
			input: "(`a`) b|cd `e` |",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenLeftParen, "("},
				{pipeline.Pos{1, 1}, pipeline.TokenString, "`a`"},
				{pipeline.Pos{1, 4}, pipeline.TokenRightParen, ")"},
				{pipeline.Pos{1, 6}, pipeline.TokenText, "b"},
				{pipeline.Pos{1, 7}, pipeline.TokenPipe, "|"},
				{pipeline.Pos{1, 8}, pipeline.TokenText, "cd"},
				{pipeline.Pos{1, 11}, pipeline.TokenString, "`e`"},
				{pipeline.Pos{1, 15}, pipeline.TokenPipe, "|"},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.input, func(t *testing.T) {
			require := require.New(t)

			l := pipeline.NewLexer(bytes.NewReader([]byte(tc.input)))
			for _, expected := range tc.expected {
				actual, err := l.Lex()
				require.NoError(err)
				require.Equal(expected, actual)
			}

			eof, err := l.Lex()
			require.NoError(err)
			require.Equal(pipeline.TokenEOF, eof.Token)
		})
	}
}

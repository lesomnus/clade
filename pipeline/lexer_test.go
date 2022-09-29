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
			input: "\"()|\"",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenText, "\"()|\""},
			},
		},
		{
			input: "()|foo\"bar\"",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenLeftParen, "("},
				{pipeline.Pos{1, 1}, pipeline.TokenRightParen, ")"},
				{pipeline.Pos{1, 2}, pipeline.TokenPipe, "|"},
				{pipeline.Pos{1, 3}, pipeline.TokenText, "foo"},
				{pipeline.Pos{1, 6}, pipeline.TokenText, "\"bar\""},
			},
		},
		{
			input: "(foo|\"bar\")",
			expected: []pipeline.TokenItem{
				{pipeline.Pos{1, 0}, pipeline.TokenLeftParen, "("},
				{pipeline.Pos{1, 1}, pipeline.TokenText, "foo"},
				{pipeline.Pos{1, 4}, pipeline.TokenPipe, "|"},
				{pipeline.Pos{1, 5}, pipeline.TokenText, "\"bar\""},
				{pipeline.Pos{1, 10}, pipeline.TokenRightParen, ")"},
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

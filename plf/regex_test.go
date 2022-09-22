package plf_test

import (
	"testing"

	"github.com/lesomnus/clade/plf"
	"github.com/stretchr/testify/require"
)

func TestRegex(t *testing.T) {
	require := require.New(t)

	rst, err := plf.Regex(`(\w) (?P<foo>\w) (\w) (?P<bar>\w)`, "a b c d")
	require.NoError(err)
	require.Len(rst, 1)

	require.Equal([]string{"a b c d", "a", "b", "c", "d"}, rst[0]["_"])
	require.Equal("b", rst[0]["foo"])
	require.Equal("d", rst[0]["bar"])
}

func TestRegexString(t *testing.T) {
	require := require.New(t)

	rst, err := plf.Regex(`fo`, "foo")
	require.NoError(err)
	require.Len(rst, 1)
	require.Equal("foo", rst[0].String())
}

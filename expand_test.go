package clade_test

import (
	"context"
	"testing"

	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
)

func TestRegexPopulator(t *testing.T) {
	require := require.New(t)

	named, err := clade.ParseReference(`cr.io/repo/golang:/^1\.(?P<minor>1[8-9]|[2-9][0-9])\.(?P<patch>\d+)$/`)
	require.NoError(err)

	regex_tagged, ok := named.(clade.RefNamedRegexTagged)
	require.True(ok)

	image := &clade.NamedImage{
		Tags: []string{
			"1.$minor.$patch",
			"1.$minor",
		},
		From: regex_tagged,
	}

	populate := clade.NewRegexExpander([]string{
		"1.17.0",
		"1.17.0-alpine",
		"1.17.0-alpine3.14",
		"1.17.1",
		"1.17.1-alpine",
		"1.17.1-alpine3.15",
		"1.18.0",
		"1.18.0-alpine",
		"1.18.0-alpine3.15",
		"1.18.1",
		"1.18.1-alpine",
		"1.18.1-alpine3.15",
		"1.18.9",
		"1.18.9-alpine",
		"1.18.9-alpine3.15",
		"1.18.10",
		"1.18.10-alpine",
		"1.18.10-alpine3.15",
		"1.18",
		"1.18-alpine",
		"1.18-alpine3.15",
		"1.19.0",
		"1.19.0-alpine",
		"1.19.0-alpine3.16",
		"1.19",
		"1.19-alpine",
		"1.19-alpine3.16",
	})

	populated, err := populate(context.Background(), image)
	require.NoError(err)

	actual := make([][]string, 0)
	for _, img := range populated {
		actual = append(actual, img.Tags)
	}

	expected := [][]string{
		{"1.18.0", "1.18"},
		{"1.18.1", "1.18"},
		{"1.18.9", "1.18"},
		{"1.18.10", "1.18"},
		{"1.19.0", "1.19"},
	}

	require.ElementsMatch(expected, actual)
}

package clade_test

import (
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/pl"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestImageReferenceUnmarshalYaml(t *testing.T) {
	tcs := []struct {
		desc     string
		data     string
		expected clade.ImageReference
	}{
		{
			desc: "tagged string",
			data: `cr.io/repo/name:tag`,
			expected: clade.ImageReference{
				Named: must(reference.ParseNamed("cr.io/repo/name")),
				Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "tag")))),
			},
		},
		{
			desc: "digested string",
			data: `cr.io/repo/name@algo:0123456789abcdef0123456789abcdef`,
			expected: clade.ImageReference{
				Named: must(reference.ParseNamed("cr.io/repo/name")),
				Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "algo:0123456789abcdef0123456789abcdef")))),
			},
		},
		{
			desc: "pipelined string",
			data: `cr.io/repo/name:(foo "bar")`,
			expected: clade.ImageReference{
				Named: must(reference.ParseNamed("cr.io/repo/name")),
				Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("foo", "bar")))),
			},
		},
		{
			desc: "tagged map",
			data: `{"name": "cr.io/repo/name", "tag": "tag"}`,
			expected: clade.ImageReference{
				Named: must(reference.ParseNamed("cr.io/repo/name")),
				Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "tag")))),
			},
		},
		{
			desc: "digested map",
			data: `{"name": "cr.io/repo/name", "tag": "algo:0123456789abcdef0123456789abcdef"}`,
			expected: clade.ImageReference{
				Named: must(reference.ParseNamed("cr.io/repo/name")),
				Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "algo:0123456789abcdef0123456789abcdef")))),
			},
		},
		{
			desc: "pipelined map",
			data: `{"name": "cr.io/repo/name", "tag": "(foo \"bar\")"}`,
			expected: clade.ImageReference{
				Named: must(reference.ParseNamed("cr.io/repo/name")),
				Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("foo", "bar")))),
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			var actual clade.ImageReference
			err := yaml.Unmarshal([]byte(tc.data), &actual)
			require.NoError(err)
			require.Equal(tc.expected, actual)
		})
	}

	t.Run("fails is", func(t *testing.T) {
		tcs := []struct {
			desc string
			data string
			msgs []string
		}{
			{
				desc: "not a string or a map",
				data: "[]",
				msgs: []string{"string", "map"},
			},
			{
				desc: "invalid name in string format",
				data: "name:tag",
				msgs: []string{"canonical"},
			},
			{
				desc: "invalid tag in string format",
				data: "cr.io/repo/name:foo bar",
				msgs: []string{"tag format"},
			},
			{
				desc: "invalid digest in string format",
				data: "cr.io/repo/name@foo",
				msgs: []string{"digest format"},
			},
			{
				desc: "invalid name in map format",
				data: `{"name": "name", "tag": "tag"}`,
				msgs: []string{"canonical"},
			},
			{
				desc: "invalid tag in map format",
				data: `{"name": "cr.io/repo/name", "tag": "foo bar"}`,
				msgs: []string{"tag format"},
			},
			{
				desc: "invalid digest in map format",
				data: `{"name": "cr.io/repo/name", "tag": "algo:short"}`,
				msgs: []string{"digest format"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var actual clade.ImageReference
				err := yaml.Unmarshal([]byte(tc.data), &actual)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

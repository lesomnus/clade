package clade_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/pl"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func TestBaseImageUnamarshalYaml(t *testing.T) {
	ref := &clade.ImageReference{
		Named: must(reference.ParseNamed("cr.io/repo/name")),
		Tag:   (*clade.Pipeline)(pl.NewPl(must(pl.NewFn("pass", "tag")))),
	}

	tcs := []struct {
		desc     string
		data     string
		expected clade.BaseImage
	}{
		{
			desc: "string",
			data: "cr.io/repo/name:tag",
			expected: clade.BaseImage{
				Primary:     ref,
				Secondaries: nil,
			},
		},
		{
			desc: "map",
			data: "{name: cr.io/repo/name, tags: tag}",
			expected: clade.BaseImage{
				Primary:     ref,
				Secondaries: nil,
			},
		},
		{
			desc: "field `with`",
			data: `
name: cr.io/repo/name
tags: tag
with:
  - cr.io/repo/name:tag
  - name: cr.io/repo/name
    tag: tag
`,
			expected: clade.BaseImage{
				Primary:     ref,
				Secondaries: []*clade.ImageReference{ref, ref},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			var actual clade.BaseImage
			err := yaml.Unmarshal([]byte(tc.data), &actual)
			require.NoError(err)
			require.Equal(tc.expected, actual)
		})
	}

	t.Run("fails if", func(t *testing.T) {
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
				desc: "invalid ImageReference type",
				data: "{name: {}}",
				msgs: []string{"map into string"},
			},
			{
				desc: "invalid ImageReference format",
				data: "{name: cr.io/repo/name, tags: (foo bar)}",
				msgs: []string{"bar"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var pipeline clade.BaseImage
				err := yaml.Unmarshal([]byte(tc.data), &pipeline)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

// func TestImageUnmarshalYaml(t *testing.T) {
// 	t.Run("must be object", func(t *testing.T) {
// 		var image clade.Image
// 		err := yaml.Unmarshal([]byte("foo"), &image)
// 		require.ErrorContains(t, err, "str")
// 		require.ErrorContains(t, err, "into")
// 	})

// 	t.Run(".platform", func(t *testing.T) {
// 		t.Run("is boolean expression", func(t *testing.T) {
// 			var image clade.Image
// 			err := yaml.Unmarshal([]byte("platform: t & f"), &image)
// 			require.NoError(t, err)
// 		})

// 		t.Run("can be empty", func(t *testing.T) {
// 			var image clade.Image
// 			err := yaml.Unmarshal([]byte("platform: "), &image)
// 			require.NoError(t, err)
// 		})

// 		t.Run("fails if", func(t *testing.T) {
// 			t.Run("not a string", func(t *testing.T) {
// 				var image clade.Image
// 				err := yaml.Unmarshal([]byte("{platform: {}}"), &image)
// 				require.ErrorContains(t, err, "map into string")
// 			})

// 			t.Run("not a valid boolean expression", func(t *testing.T) {
// 				var image clade.Image
// 				err := yaml.Unmarshal([]byte("{platform: foo %% bar}"), &image)
// 				require.ErrorContains(t, err, "%")
// 			})
// 		})
// 	})
// }

func TestResolvedImage(t *testing.T) {
	t.Run("Tagged", func(t *testing.T) {
		named, err := reference.ParseNamed("cr.io/repo/name")
		require.NoError(t, err)

		t.Run("tagged by first tag with its name", func(t *testing.T) {
			require := require.New(t)

			img := clade.ResolvedImage{
				Named: named,
				Tags:  []string{"foo", "bar"},
			}
			tagged, err := img.Tagged()
			require.NoError(err)
			require.Equal(named.Name(), tagged.Name())
			require.Equal("foo", tagged.Tag())
		})

		t.Run("fails if", func(t *testing.T) {
			t.Run("it has no tags", func(t *testing.T) {
				require := require.New(t)

				img := clade.ResolvedImage{Named: named}
				_, err := img.Tagged()
				require.ErrorContains(err, "not tagged")
			})

			t.Run("tag is invalid", func(t *testing.T) {
				require := require.New(t)

				img := clade.ResolvedImage{
					Named: named,
					Tags:  []string{"foo bar"},
				}
				_, err := img.Tagged()
				require.ErrorIs(err, reference.ErrTagInvalidFormat)
			})
		})
	})
}

func TestCalcDerefId(t *testing.T) {
	require := require.New(t)

	a := []byte{0x42, 0x08, 0xD2}
	b := []byte{0x30, 0x9A, 0xAD, 0x51}
	c := []byte{0xE0, 0x9C, 0xA8, 0x07, 0xB6}

	all := []string{
		clade.CalcDerefId(a, b, c),
		clade.CalcDerefId(a, c, b),
		clade.CalcDerefId(b, a, c),
		clade.CalcDerefId(b, c, a),
		clade.CalcDerefId(c, a, b),
		clade.CalcDerefId(c, b, a),
	}

	for _, lhs := range all {
		for _, rhs := range all {
			require.Equal(lhs, rhs)
		}
	}
}

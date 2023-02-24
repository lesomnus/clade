package clade_test

import (
	"testing"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/pl"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPipelineUnmarshalYAML(t *testing.T) {
	tcs := []struct {
		desc     string
		data     string
		expected *pl.Fn
	}{
		{
			desc:     "given string is passed if it is not in parentheses",
			data:     `foo "bar"`,
			expected: must(pl.NewFn("pass", `foo "bar"`)),
		},
		{
			desc:     "pipeline expression",
			data:     `(foo "bar")`,
			expected: must(pl.NewFn("foo", "bar")),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			var actual clade.Pipeline
			err := yaml.Unmarshal([]byte(tc.data), &actual)
			require.NoError(err)
			require.Len(actual.Funcs, 1)
			require.Equal(tc.expected, actual.Funcs[0])
		})
	}

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc string
			data string
			msgs []string
		}{
			{
				desc: "not a string",
				data: "{}",
				msgs: []string{"map into string"},
			},
			{
				desc: "invalid pipeline expression",
				data: "(foo bar)",
				msgs: []string{"bar"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var actual clade.Pipeline
				err := yaml.Unmarshal([]byte(tc.data), &actual)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

func TestBoolAlgebraUnmarshalYAML(t *testing.T) {
	t.Run("string is parsed", func(t *testing.T) {
		require := require.New(t)

		var actual clade.BoolAlgebra
		err := yaml.Unmarshal([]byte("x & y"), &actual)
		require.NoError(err)
	})

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc string
			data string
			msgs []string
		}{
			{
				desc: "not a string",
				data: "{}",
				msgs: []string{"map into string"},
			},
			{
				desc: "invalid expression",
				data: "x % y",
				msgs: []string{"%"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				var actual clade.BoolAlgebra
				err := yaml.Unmarshal([]byte(tc.data), &actual)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

package clade_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
)

func TestResolvePath(t *testing.T) {
	wd := must(os.Getwd())

	tcs := []struct {
		desc     string
		base     string
		path     string
		fallback string
		expected string
	}{
		{
			desc:     "joined path if `base` is absolute",
			base:     "/foo/bar",
			path:     "baz",
			fallback: "",
			expected: "/foo/bar/baz",
		},
		{
			desc:     "joined path with cwd if `base` is relative",
			base:     "foo/bar",
			path:     "baz",
			fallback: "",
			expected: filepath.Join(wd, "foo/bar/baz"),
		},
		{
			desc:     "joined path with cwd if `base` is empty",
			base:     "",
			path:     "baz",
			fallback: "",
			expected: filepath.Join(wd, "baz"),
		},
		{
			desc:     "`fallback` is joined with `base` if `path` is empty and `base` is absolute",
			base:     "/foo/bar",
			path:     "",
			fallback: "baz",
			expected: "/foo/bar/baz",
		},
		{
			desc:     "`fallback` is joined with `base` if `path` is empty and is joined with cwd if `base` is relative",
			base:     "foo/bar",
			path:     "",
			fallback: "baz",
			expected: filepath.Join(wd, "foo/bar/baz"),
		},
		{
			desc:     "`fallback` is joined with cwd if `base` and `path` are empty",
			base:     "",
			path:     "",
			fallback: "baz",
			expected: filepath.Join(wd, "baz"),
		},
		{
			desc:     "`fallback` is returned as is if it is absolute and `base` and `path` are empty",
			base:     "",
			path:     "",
			fallback: "/baz",
			expected: "/baz",
		},
		{
			desc:     "`path` is returned as is if it is absolute",
			base:     "foo",
			path:     "/bar",
			fallback: "baz",
			expected: "/bar",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			path, err := clade.ResolvePath(tc.base, tc.path, tc.fallback)
			require.NoError(err)
			require.Equal(tc.expected, path)
		})
	}
}

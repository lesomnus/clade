package clade_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
)

func TestDependencyTree(t *testing.T) {
	require := require.New(t)

	type Img struct {
		name string
		from string
	}

	imgs := []Img{
		// Parent-first insertion.
		{
			name: "local.io/foo/a",
			from: "remote.io/foo/a:a",
		},
		{
			name: "local.io/foo/a",
			from: "remote.io/foo/a:b",
		},
		{
			name: "local.io/foo/b",
			from: "local.io/foo/a:c",
		},

		// Child-first insertion.
		{
			name: "local.io/bar/b",
			from: "local.io/bar/a:c",
		},
		{
			name: "local.io/bar/a",
			from: "remote.io/bar/a:a",
		},
		{
			name: "local.io/bar/a",
			from: "remote.io/bar/a:b",
		},
	}

	dt := clade.NewDependencyTree()
	for _, img := range imgs {
		named, err := reference.ParseNamed(img.from)
		require.NoError(err)

		tagged, ok := named.(reference.NamedTagged)
		require.True(ok)

		primary := &clade.ImageReference{}
		err = primary.FromNameTag(tagged.Name(), tagged.Tag())
		require.NoError(err)

		dt.Insert(&clade.Image{
			Named: must(reference.ParseNamed(img.name)),
			From: clade.BaseImage{
				Primary: primary,
			},
		})
	}

	test := func(name string, child_name string) {
		require.Contains(dt.Tree, name)
		require.Len(dt.Tree[name].Children, 1)
		require.Contains(dt.Tree[name].Children, child_name)
	}

	test("remote.io/foo/a", "local.io/foo/a")
	test("local.io/foo/a", "local.io/foo/b")

	test("remote.io/bar/a", "local.io/bar/a")
	test("local.io/bar/a", "local.io/bar/b")
}

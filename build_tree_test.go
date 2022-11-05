package clade_test

import (
	"fmt"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
	"github.com/stretchr/testify/require"
)

func TestBuildTree(t *testing.T) {

	type Img struct {
		name string
		tags []string
		from string
	}

	imgs := []Img{
		// Parent-first insertion.
		{
			name: "local.io/foo/a",
			tags: []string{"a1", "a2", "a3"},
			from: "remote.io/foo/a:a",
		},
		{
			name: "local.io/foo/a",
			tags: []string{"b1", "b2"},
			from: "remote.io/foo/a:b",
		},
		{
			name: "local.io/foo/b",
			tags: []string{"c1", "c2", "c3", "c4"},
			from: "local.io/foo/a:c",
		},
		{
			name: "local.io/foo/b",
			tags: []string{"bb1", "bb2", "bb3", "bb4", "bb5"},
			from: "local.io/foo/a:a1",
		},

		// Child-first insertion.
		{
			name: "local.io/bar/b",
			tags: []string{"bb1", "bb2", "bb3", "bb4", "bb5"},
			from: "local.io/bar/a:a1",
		},
		{
			name: "local.io/bar/b",
			tags: []string{"c1", "c2", "c3", "c4"},
			from: "local.io/bar/a:c",
		},
		{
			name: "local.io/bar/a",
			tags: []string{"a1", "a2", "a3"},
			from: "remote.io/bar/a:a",
		},
		{
			name: "local.io/bar/a",
			tags: []string{"b1", "b2"},
			from: "remote.io/bar/a:b",
		},
	}

	bt := clade.NewBuildTree()
	for _, img := range imgs {
		bt.Insert(&clade.ResolvedImage{
			Named: must(reference.ParseNamed(img.name)),
			From:  must(clade.ParseRefNamedTagged(img.from)),
			Tags:  img.tags,
		})
	}

	t.Run("Insert", func(t *testing.T) {
		require := require.New(t)

		test := func(name string) {
			require.Contains(bt.Tree, fmt.Sprintf("remote.io/%s/a:a", name))
			require.Len(bt.Tree[fmt.Sprintf("remote.io/%s/a:a", name)].Children, 3)
			require.Nil(bt.Tree[fmt.Sprintf("remote.io/%s/a:a", name)].Parent)

			require.Contains(bt.Tree, fmt.Sprintf("remote.io/%s/a:b", name))
			require.Len(bt.Tree[fmt.Sprintf("remote.io/%s/a:b", name)].Children, 2)
			require.Nil(bt.Tree[fmt.Sprintf("remote.io/%s/a:b", name)].Parent)

			require.Contains(bt.Tree, fmt.Sprintf("local.io/%s/a:c", name))
			require.Len(bt.Tree[fmt.Sprintf("local.io/%s/a:c", name)].Children, 4)
			require.Nil(bt.Tree[fmt.Sprintf("local.io/%s/a:c", name)].Parent)

			require.Contains(bt.Tree, fmt.Sprintf("local.io/%s/a:a1", name))
			require.Len(bt.Tree[fmt.Sprintf("local.io/%s/a:a1", name)].Children, 5)
			require.NotNil(bt.Tree[fmt.Sprintf("local.io/%s/a:a1", name)].Parent)
			require.Equal(bt.Tree[fmt.Sprintf("local.io/%s/a:a1", name)].Parent, bt.Tree[fmt.Sprintf("remote.io/%s/a:a", name)])

			require.Contains(bt.Tree, fmt.Sprintf("local.io/%s/b:bb1", name))
			require.Len(bt.Tree[fmt.Sprintf("local.io/%s/b:bb1", name)].Children, 0)
			require.NotNil(bt.Tree[fmt.Sprintf("local.io/%s/b:bb1", name)].Parent)
			require.Equal(bt.Tree[fmt.Sprintf("local.io/%s/b:bb1", name)].Parent, bt.Tree[fmt.Sprintf("local.io/%s/a:a1", name)])

		}

		test("foo")
		test("bar")
	})

	t.Run("Tags by name", func(t *testing.T) {
		require := require.New(t)

		require.ElementsMatch([]string{"a1", "a2", "a3", "b1", "b2"}, bt.TagsByName["local.io/foo/a"])
		require.ElementsMatch([]string{"a1", "a2", "a3", "b1", "b2"}, bt.TagsByName["local.io/bar/a"])
		require.ElementsMatch([]string{"c1", "c2", "c3", "c4", "bb1", "bb2", "bb3", "bb4", "bb5"}, bt.TagsByName["local.io/foo/b"])
		require.ElementsMatch([]string{"c1", "c2", "c3", "c4", "bb1", "bb2", "bb3", "bb4", "bb5"}, bt.TagsByName["local.io/bar/b"])
	})

	t.Run("Insert fails if insert image with invalid tag", func(t *testing.T) {
		require := require.New(t)

		err := bt.Insert(&clade.ResolvedImage{
			Named: must(reference.ParseNamed("cr.io/foo/bar")),
			Tags:  []string{"John Wick"},
			From:  must(clade.ParseRefNamedTagged("hub.io/foo/bar:baz")),
		})

		require.ErrorContains(err, "invalid")
	})

	t.Run("Walk", func(t *testing.T) {
		require := require.New(t)

		cnt := map[string]int{}
		bt.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
			cnt[node.Value.Name()]++
			return nil
		})

		// local.io/foo/a:{a1, a2, ...}
		// local.io/foo/a:{b1, b2, ...}
		// local.io/foo/a:c
		require.Contains(cnt, "local.io/foo/a")
		require.Equal(cnt["local.io/foo/a"], 3)

		require.Contains(cnt, "local.io/foo/b")
		require.Equal(cnt["local.io/foo/b"], 2)
	})
}

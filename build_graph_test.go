package clade_test

import (
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
)

func TestBuildGraph(t *testing.T) {
	//              origin/bar:bar -| repo/bar:c
	//                              | repo/bar:d
	//                             /            \
	// origin/foo:foo -| repo/foo:a              \
	//                 | repo/foo:b --------------| repo/baz:e
	//                                           /
	//                             origin/baz:baz

	origin_foo := must(reference.Parse("cr.io/origin/foo:foo")).(reference.NamedTagged)
	origin_bar := must(reference.Parse("cr.io/origin/bar:bar")).(reference.NamedTagged)
	origin_baz := must(reference.Parse("cr.io/origin/baz:baz")).(reference.NamedTagged)

	foo := &clade.ResolvedImage{
		Named: must(reference.ParseNamed("cr.io/repo/foo")),
		Tags:  []string{"a", "b"},
		From: &clade.ResolvedBaseImage{
			Primary: clade.ResolvedImageReference{NamedTagged: origin_foo},
		},
	}
	bar := &clade.ResolvedImage{
		Named: must(reference.ParseNamed("cr.io/repo/bar")),
		Tags:  []string{"c", "d"},
		From: &clade.ResolvedBaseImage{
			Primary: clade.ResolvedImageReference{NamedTagged: origin_bar},
			Secondaries: []clade.ResolvedImageReference{
				{NamedTagged: must(reference.Parse("cr.io/repo/foo:a")).(reference.NamedTagged)},
			},
		},
	}
	baz := &clade.ResolvedImage{
		Named: must(reference.ParseNamed("cr.io/repo/baz")),
		Tags:  []string{"e"},
		From: &clade.ResolvedBaseImage{
			Primary: clade.ResolvedImageReference{NamedTagged: origin_baz},
			Secondaries: []clade.ResolvedImageReference{
				{NamedTagged: must(reference.Parse("cr.io/repo/foo:b")).(reference.NamedTagged)},
				{NamedTagged: must(reference.Parse("cr.io/repo/bar:d")).(reference.NamedTagged)},
			},
		},
	}

	test_graph := func(t *testing.T, graph *clade.BuildGraph) {
		require := require.New(t)

		roots := graph.Roots()
		require.Len(roots, 3)
		require.ElementsMatch(
			[]string{origin_foo.String(), origin_bar.String(), origin_baz.String()},
			[]string{roots[0].Key(), roots[1].Key(), roots[2].Key()},
		)

		tags, ok := graph.TagsByName(foo.Named)
		require.True(ok)
		require.ElementsMatch([]string{"a", "b"}, tags)

		tags, ok = graph.TagsByName(bar.Named)
		require.True(ok)
		require.ElementsMatch([]string{"c", "d"}, tags)

		tags, ok = graph.TagsByName(baz.Named)
		require.True(ok)
		require.ElementsMatch([]string{"e"}, tags)

		for _, node := range roots {
			switch node.Key() {
			case origin_foo.String():
				require.Len(node.Prev, 0)
				require.Len(node.Next, 2)
				require.Contains(node.Next, "cr.io/repo/foo:a")
				require.Contains(node.Next, "cr.io/repo/foo:b")

				foo_next_a := node.Next["cr.io/repo/foo:a"]
				require.Len(foo_next_a.Prev, 1)
				require.Contains(foo_next_a.Prev, "cr.io/origin/foo:foo")
				require.Len(foo_next_a.Next, 2)
				require.Contains(foo_next_a.Next, "cr.io/repo/bar:c")
				require.Contains(foo_next_a.Next, "cr.io/repo/bar:d")

				foo_next_b := node.Next["cr.io/repo/foo:b"]
				require.Len(foo_next_b.Prev, 1)
				require.Contains(foo_next_b.Prev, "cr.io/origin/foo:foo")
				require.Len(foo_next_b.Next, 1)
				require.Contains(foo_next_b.Next, "cr.io/repo/baz:e")

			case origin_bar.String():
				require.Len(node.Prev, 0)
				require.Len(node.Next, 2)
				require.Contains(node.Next, "cr.io/repo/bar:c")
				require.Contains(node.Next, "cr.io/repo/bar:d")

				bar_next_a := node.Next["cr.io/repo/bar:c"]
				require.Len(bar_next_a.Prev, 2)
				require.Contains(bar_next_a.Prev, "cr.io/origin/bar:bar")
				require.Contains(bar_next_a.Prev, "cr.io/repo/foo:a")
				require.Len(bar_next_a.Next, 0)

				bar_next_b := node.Next["cr.io/repo/bar:d"]
				require.Len(bar_next_b.Prev, 2)
				require.Contains(bar_next_b.Prev, "cr.io/origin/bar:bar")
				require.Contains(bar_next_b.Prev, "cr.io/repo/foo:a")
				require.Len(bar_next_b.Next, 1)
				require.Contains(bar_next_b.Next, "cr.io/repo/baz:e")

			case origin_baz.String():
				require.Len(node.Prev, 0)
				require.Len(node.Next, 1)
				require.Contains(node.Next, "cr.io/repo/baz:e")

				baz_next_b := node.Next["cr.io/repo/baz:e"]
				require.Len(baz_next_b.Prev, 3)
				require.Contains(baz_next_b.Prev, "cr.io/origin/baz:baz")
				require.Contains(baz_next_b.Prev, "cr.io/repo/foo:b")
				require.Contains(baz_next_b.Prev, "cr.io/repo/bar:d")
				require.Len(baz_next_b.Next, 0)

			default:
				require.Fail("unexpected key", node.Key())
			}
		}
	}

	t.Run("in order", func(t *testing.T) {
		graph := clade.NewBuildGraph()
		graph.Put(foo)
		graph.Put(bar)
		graph.Put(baz)

		test_graph(t, graph)
	})

	t.Run("in revers order", func(t *testing.T) {
		graph := clade.NewBuildGraph()
		graph.Put(baz)
		graph.Put(bar)
		graph.Put(foo)

		test_graph(t, graph)
	})

	t.Run("from middle", func(t *testing.T) {
		graph := clade.NewBuildGraph()
		graph.Put(bar)
		graph.Put(foo)
		graph.Put(baz)

		test_graph(t, graph)
	})

	t.Run("TagsByName will return false if there is no image such name", func(t *testing.T) {
		require := require.New(t)
		graph := clade.NewBuildGraph()
		_, ok := graph.TagsByName(foo.Named)
		require.False(ok)
	})

	t.Run("fails if", func(t *testing.T) {
		t.Run("it has no tags", func(t *testing.T) {
			require := require.New(t)

			graph := clade.NewBuildGraph()
			_, err := graph.Put(&clade.ResolvedImage{
				Named: must(reference.ParseNamed("cr.io/repo/foo")),
				From: &clade.ResolvedBaseImage{
					Primary: clade.ResolvedImageReference{NamedTagged: origin_foo},
				},
			})
			require.ErrorContains(err, "no tags")
		})

		t.Run("tag is invalid", func(t *testing.T) {
			require := require.New(t)

			graph := clade.NewBuildGraph()
			_, err := graph.Put(&clade.ResolvedImage{
				Named: must(reference.ParseNamed("cr.io/repo/foo")),
				Tags:  []string{"a b"},
				From: &clade.ResolvedBaseImage{
					Primary: clade.ResolvedImageReference{NamedTagged: origin_foo},
				},
			})
			require.ErrorContains(err, "a b")
		})

		t.Run("reference is duplicated with another image", func(t *testing.T) {
			require := require.New(t)

			graph := clade.NewBuildGraph()
			image := &clade.ResolvedImage{
				Named: must(reference.ParseNamed("cr.io/repo/foo")),
				Tags:  []string{"a"},
				From: &clade.ResolvedBaseImage{
					Primary: clade.ResolvedImageReference{NamedTagged: origin_foo},
				},
			}

			_, err := graph.Put(image)
			require.NoError(err)

			_, err = graph.Put(image)
			require.NoError(err)

			// Same content but different instance.
			_, err = graph.Put(&clade.ResolvedImage{
				Named: must(reference.ParseNamed("cr.io/repo/foo")),
				Tags:  []string{"a"},
				From: &clade.ResolvedBaseImage{
					Primary: clade.ResolvedImageReference{NamedTagged: origin_foo},
				},
			})
			require.ErrorContains(err, "duplicate")
		})
	})
}

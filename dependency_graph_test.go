package clade_test

import (
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
)

func TestDependencyGraph(t *testing.T) {
	//          origin/bar - repo/bar
	//                      /        \
	// origin/foo - repo/foo -------- repo/baz
	//                               /
	//                     origin/baz

	foo_origin := &clade.ImageReference{Named: must(reference.ParseNamed("cr.io/origin/foo"))}
	bar_origin := &clade.ImageReference{Named: must(reference.ParseNamed("cr.io/origin/bar"))}
	baz_origin := &clade.ImageReference{Named: must(reference.ParseNamed("cr.io/origin/baz"))}

	foo := &clade.Image{
		Named: must(reference.ParseNamed("cr.io/repo/foo")),
		From: clade.BaseImage{
			Primary: foo_origin,
		},
	}
	bar := &clade.Image{
		Named: must(reference.ParseNamed("cr.io/repo/bar")),
		From: clade.BaseImage{
			Primary: bar_origin,
			Secondaries: []*clade.ImageReference{
				{Named: foo.Named},
			},
		},
	}
	baz := &clade.Image{
		Named: must(reference.ParseNamed("cr.io/repo/baz")),
		From: clade.BaseImage{
			Primary: baz_origin,
			Secondaries: []*clade.ImageReference{
				{Named: foo.Named},
				{Named: bar.Named},
				{Named: bar.Named},
			},
		},
	}

	test_graph := func(t *testing.T, graph *clade.DependencyGraph) {
		require := require.New(t)

		roots := graph.Roots()
		require.Len(roots, 3)
		require.ElementsMatch(
			[]string{foo_origin.Name(), bar_origin.Name(), baz_origin.Name()},
			[]string{roots[0].Key(), roots[1].Key(), roots[2].Key()},
		)

		for _, node := range roots {
			switch node.Key() {
			case foo_origin.Name():
				require.Len(node.Prev, 0)
				require.Len(node.Next, 1)
				require.Contains(node.Next, foo.Name())

				foo_next := node.Next[foo.Name()]
				require.Len(foo_next.Value, 1)
				require.Equal(foo, foo_next.Value[0])
				require.Len(foo_next.Prev, 1)
				require.Len(foo_next.Next, 2)
				require.Contains(foo_next.Next, bar.Name())
				require.Contains(foo_next.Next, baz.Name())

			case bar_origin.Name():
				require.Len(node.Prev, 0)
				require.Len(node.Next, 1)
				require.Contains(node.Next, bar.Name())

				bar_next := node.Next[bar.Name()]
				require.Len(bar_next.Value, 1)
				require.Equal(bar, bar_next.Value[0])
				require.Len(bar_next.Prev, 2)
				require.Len(bar_next.Next, 1)
				require.Contains(bar_next.Next, baz.Name())

			case baz_origin.Name():
				require.Len(node.Prev, 0)
				require.Len(node.Next, 1)
				require.Contains(node.Next, baz.Name())

				baz_next := node.Next[baz.Name()]
				require.Len(baz_next.Value, 1)
				require.Equal(baz, baz_next.Value[0])
				require.Len(baz_next.Prev, 3)
				require.Len(baz_next.Next, 0)

			default:
				require.Fail("unexpected key", node.Key())
			}
		}
	}

	t.Run("in order", func(t *testing.T) {
		graph := clade.NewDependencyGraph()
		graph.Put(foo)
		graph.Put(bar)
		graph.Put(baz)

		test_graph(t, graph)
	})

	t.Run("in revers order", func(t *testing.T) {
		graph := clade.NewDependencyGraph()
		graph.Put(baz)
		graph.Put(bar)
		graph.Put(foo)

		test_graph(t, graph)
	})

	t.Run("from middle", func(t *testing.T) {
		graph := clade.NewDependencyGraph()
		graph.Put(bar)
		graph.Put(foo)
		graph.Put(baz)

		test_graph(t, graph)
	})
}

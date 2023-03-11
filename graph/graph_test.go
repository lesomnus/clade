package graph_test

import (
	"testing"

	"github.com/lesomnus/clade/graph"
	"github.com/stretchr/testify/require"
)

func TestGraph(t *testing.T) {
	require := require.New(t)

	g := graph.NewGraph[string]()

	a := g.Put("a", "a")
	b := g.Put("b", "b")
	c := g.Put("c", "c")
	d := g.Put("d", "d")
	e := g.Put("e", "e")

	// a - c - d
	//   /   \
	// b       e
	a.ConnectTo(c)
	b.ConnectTo(c)
	c.ConnectTo(d)
	c.ConnectTo(e)

	roots := g.Roots()
	require.Len(roots, 2)
	require.ElementsMatch([]*graph.Node[string]{a, b}, roots)

	leaves := g.Leaves()
	require.Len(leaves, 2)
	require.ElementsMatch([]*graph.Node[string]{d, e}, leaves)

	lookuped_c, ok := g.Get("c")
	require.True(ok)
	require.Equal(c, lookuped_c)

	lookuped_c = g.GetOrPut("c", "z")
	require.Equal(c, lookuped_c)

	lookuped_z := g.GetOrPut("z", "z")
	require.Equal("z", lookuped_z.Key())
	require.Equal("z", lookuped_z.Value)
}

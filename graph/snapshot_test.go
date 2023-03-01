package graph_test

import (
	"testing"

	"github.com/lesomnus/clade/graph"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func TestSnapshot(t *testing.T) {
	require := require.New(t)

	g := graph.NewGraph[string]()

	a := g.Put("a", "a")
	b := g.Put("b", "b")
	c := g.Put("c", "c")
	d := g.Put("d", "d")
	e := g.Put("e", "e")
	f := g.Put("f", "f")
	h := g.Put("h", "h")
	i := g.Put("i", "i")
	j := g.Put("j", "j")

	// a - b - c
	// d - e - f
	//   /   \
	// h      \
	// i ----- j
	a.ConnectTo(b)
	b.ConnectTo(c)
	d.ConnectTo(e)
	e.ConnectTo(f)
	e.ConnectTo(j)
	h.ConnectTo(e)
	i.ConnectTo(j)

	snapshot := g.Snapshot(nil)
	require.ElementsMatch([]string{"a", "b", "c", "d", "e", "f", "h", "i", "j"}, maps.Keys(snapshot))
	require.Equal(uint(0), snapshot["a"].Level)
	require.Equal(uint(1), snapshot["b"].Level)
	require.Equal(uint(2), snapshot["c"].Level)
	require.Equal(uint(0), snapshot["d"].Level)
	require.Equal(uint(1), snapshot["e"].Level)
	require.Equal(uint(2), snapshot["f"].Level)
	require.Equal(uint(0), snapshot["h"].Level)
	require.Equal(uint(0), snapshot["i"].Level)
	require.Equal(uint(2), snapshot["j"].Level)

	require.EqualValues(snapshot["a"].Group, snapshot["b"].Group)
	require.EqualValues(snapshot["a"].Group, snapshot["c"].Group)
	require.EqualValues(snapshot["b"].Group, snapshot["c"].Group)

	require.Subset(maps.Keys(snapshot["e"].Group), maps.Keys(snapshot["f"].Group))
	require.Subset(maps.Keys(snapshot["e"].Group), maps.Keys(snapshot["j"].Group))
	require.Subset(maps.Keys(snapshot["d"].Group), maps.Keys(snapshot["e"].Group))
	require.Subset(maps.Keys(snapshot["h"].Group), maps.Keys(snapshot["e"].Group))
	require.Subset(maps.Keys(snapshot["i"].Group), maps.Keys(snapshot["j"].Group))
}

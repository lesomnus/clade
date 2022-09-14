package tree_test

import (
	"testing"

	"github.com/lesomnus/clade/tree"
	"github.com/stretchr/testify/require"
)

func TestTree(t *testing.T) {
	require := require.New(t)

	tr := make(tree.Tree[int])

	tr.Insert("mammalia", "whale", 2)
	tr.Insert("mammalia", "human", 2)
	mammalia := tr.Insert("animal", "mammalia", 1)

	tr.Insert("amphibians", "frog", 2)
	tr.Insert("amphibians", "salamander", 2)
	amphibians := tr.Insert("animal", "amphibians", 1)

	root := tr.AsNode()
	require.Len(root.Children, 1)

	animal, ok := root.Children["animal"]
	require.True(ok)
	require.Equal(animal, mammalia.Parent)
	require.Equal(animal, amphibians.Parent)
	require.Contains(animal.Children, "mammalia")
	require.Contains(animal.Children, "amphibians")

	// Note that root node is not visited.
	root.Walk(func(level int, name string, node *tree.Node[int]) error {
		require.Equal(level, node.Value)

		return nil
	})
}

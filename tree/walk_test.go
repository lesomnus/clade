package tree_test

import (
	"io"
	"testing"

	"github.com/lesomnus/clade/tree"
	"github.com/stretchr/testify/require"
)

func TestWalkContinue(t *testing.T) {
	require := require.New(t)

	tr := make(tree.Tree[bool])
	tr.Insert("_", "a", true)
	tr.Insert("_", "b", false)
	tr.Insert("a", "x", false)
	tr.Insert("b", "y", false)

	vs := make([]string, 0)

	tr.AsNode().Walk(func(level int, name string, node *tree.Node[bool]) error {
		if node.Value {
			return tree.WalkContinue
		}

		vs = append(vs, name)
		return nil
	})

	require.ElementsMatch([]string{"_", "b", "y"}, vs)
}

func TestWalkBreak(t *testing.T) {
	require := require.New(t)

	tr := make(tree.Tree[bool])
	tr.Insert("_", "a", false)
	tr.Insert("_", "b", false)
	tr.Insert("a", "x", true)
	tr.Insert("x", "1", false)
	tr.Insert("b", "y", true)
	tr.Insert("y", "2", false)

	vs := make([]string, 0)

	tr.AsNode().Walk(func(level int, name string, node *tree.Node[bool]) error {
		if node.Value {
			return tree.WalkBreak
		}

		vs = append(vs, name)
		return nil
	})

	require.ElementsMatch([]string{"_", "a", "b"}, vs)
}

func TestWalkError(t *testing.T) {
	require := require.New(t)

	tr := make(tree.Tree[bool])
	tr.Insert("_", "a", true)
	tr.Insert("_", "b", true)

	vs := make([]string, 0)

	err := tr.AsNode().Walk(func(level int, name string, node *tree.Node[bool]) error {
		if node.Value {
			return io.EOF
		}

		vs = append(vs, name)
		return nil
	})
	require.ErrorIs(err, io.EOF)
	require.ElementsMatch(vs, []string{"_"})
}

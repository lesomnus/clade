package clade_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
)

func TestBuildTreeInsert(t *testing.T) {
	require := require.New(t)

	parent_name, err := reference.ParseNamed("cr1.io/repo/gcc:12")
	require.NoError(err)

	child_name, err := reference.ParseNamed("cr2.io/repo/boost:1.79")
	require.NoError(err)

	grandchild_name, err := reference.ParseNamed("cr3.io/repo/pcl")
	require.NoError(err)

	child_img := &clade.NamedImage{
		Name: child_name,
		Tags: []string{"1.79.0", "1.79"},
		From: parent_name.(reference.NamedTagged),

		Dockerfile:  "/path/to/child/dockerfile",
		ContextPath: "/path/to/child",
	}

	grandchild_img := &clade.NamedImage{
		Name: grandchild_name,
		Tags: []string{"1.12.1", "1.12"},
		From: child_name.(reference.NamedTagged),

		Dockerfile:  "/path/to/grandchild/dockerfile",
		ContextPath: "/path/to/grandchild",
	}

	bt := make(clade.BuildTree)

	err = bt.Insert(grandchild_img)
	require.NoError(err)
	{
		grandchild0, err := reference.WithTag(grandchild_name, "1.12.1")
		require.NoError(err)

		grandchild1, err := reference.WithTag(grandchild_name, "1.12")
		require.NoError(err)

		child_node, ok := bt[child_name.String()]
		require.True(ok)

		grandchild0_node, ok := bt[grandchild0.String()]
		require.True(ok)

		grandchild1_node, ok := bt[grandchild1.String()]
		require.True(ok)

		require.Equal(child_node, grandchild0_node.Parent)
		require.Equal(child_node, grandchild1_node.Parent)

		require.Equal(grandchild0_node, child_node.Children[grandchild0.String()])
		require.Equal(grandchild1_node, child_node.Children[grandchild1.String()])

		// Child node is not inserted yet but the node is reserved.
		require.Empty(child_node.BuildContext.NamedImage.Dockerfile)
		require.Empty(child_node.BuildContext.NamedImage.ContextPath)
	}

	err = bt.Insert(child_img)
	require.NoError(err)

	{
		child0, err := reference.WithTag(child_name, "1.79.0")
		require.NoError(err)

		child1, err := reference.WithTag(child_name, "1.79")
		require.NoError(err)

		parent_node, ok := bt[parent_name.String()]
		require.True(ok)

		child0_node, ok := bt[child0.String()]
		require.True(ok)

		child1_node, ok := bt[child1.String()]
		require.True(ok)

		require.Equal(parent_node, child0_node.Parent)
		require.Equal(parent_node, child1_node.Parent)

		require.Equal(child0_node, parent_node.Children[child0.String()])
		require.Equal(child1_node, parent_node.Children[child1.String()])

		// Reserved node is filled with content.
		require.Equal(child_img.Dockerfile, child0_node.BuildContext.NamedImage.Dockerfile)
		require.Equal(child_img.ContextPath, child0_node.BuildContext.NamedImage.ContextPath)
		require.Equal(child_img.Dockerfile, child1_node.BuildContext.NamedImage.Dockerfile)
		require.Equal(child_img.ContextPath, child1_node.BuildContext.NamedImage.ContextPath)
	}
}

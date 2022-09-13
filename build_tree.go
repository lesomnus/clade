package clade

import (
	"errors"
	"fmt"

	"github.com/distribution/distribution/reference"
	"golang.org/x/exp/slices"
)

var (
	WalkContinue = errors.New("") //lint:ignore ST1012 Special error for indicating
	WalkBreak    = errors.New("") //lint:ignore ST1012 Special error for indicating
)

type BuildContext struct {
	NamedImage *NamedImage
}

type BuildTreeNode struct {
	Parent       *BuildTreeNode
	Children     map[string]*BuildTreeNode
	BuildContext BuildContext
}

type BuildTreeWalker func(level int, node *BuildTreeNode) error

func walk(node *BuildTreeNode, level int, fn BuildTreeWalker) error {
	visited := make([]*BuildTreeNode, 0, len(node.Children))
	for _, child := range node.Children {
		if slices.Contains(visited, child) {
			continue
		} else {
			visited = append(visited, child)
		}

		if err := fn(level, child); err != nil {
			if errors.Is(err, WalkContinue) {
				continue
			} else if errors.Is(err, WalkBreak) {
				return nil
			} else {
				return err
			}
		}

		if err := walk(child, level+1, fn); err != nil {
			return err
		}
	}

	return nil
}

func (n *BuildTreeNode) Walk(fn BuildTreeWalker) error {
	return walk(n, 0, fn)
}

func (n *BuildTreeNode) IsRoot() bool {
	return n.Parent == nil
}

type BuildTree map[string]*BuildTreeNode

func (bt BuildTree) AsNode() *BuildTreeNode {
	children := make(map[string]*BuildTreeNode)
	for name, node := range bt {
		if !node.IsRoot() {
			continue
		}

		children[name] = node
	}

	return &BuildTreeNode{
		Parent:       nil,
		Children:     children,
		BuildContext: BuildContext{NamedImage: nil},
	}
}

func (bt BuildTree) Walk(fn BuildTreeWalker) error {
	return walk(bt.AsNode(), 0, fn)
}

func (bt BuildTree) Insert(image *NamedImage) error {
	parent_name := image.From.String()
	parent, ok := bt[parent_name]
	if !ok {
		parent = &BuildTreeNode{
			Parent:   nil,
			Children: make(map[string]*BuildTreeNode),
			BuildContext: BuildContext{
				NamedImage: &NamedImage{
					Name: image.From,
					Tags: []string{image.From.Tag()},
				},
			},
		}

		bt[parent_name] = parent
	}

	child_names := make([]reference.NamedTagged, len(image.Tags))
	for i, tag := range image.Tags {
		name_tagged, err := reference.WithTag(image.Name, tag)
		if err != nil {
			return fmt.Errorf("failed to tag %s with %s: %w", image.Name.String(), tag, err)
		}

		child_names[i] = name_tagged
	}

	// Check if there is reserved node with shared tag of given image.
	var child_node *BuildTreeNode
	for _, child_name := range child_names {
		if c, ok := bt[child_name.String()]; ok {
			child_node = c
			break
		}
	}

	if child_node == nil {
		child_node = &BuildTreeNode{
			Parent:       parent,
			Children:     make(map[string]*BuildTreeNode),
			BuildContext: BuildContext{NamedImage: image},
		}
	} else {
		// Fill data to reserved node.
		child_node.Parent = parent
		child_node.BuildContext.NamedImage = image
	}

	for _, child_name := range child_names {
		name := child_name.String()
		bt[name] = child_node
		parent.Children[name] = child_node
	}

	return nil
}

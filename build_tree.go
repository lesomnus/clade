package clade

import (
	"github.com/docker/distribution/reference"
	"github.com/lesomnus/clade/tree"
)

type BuildTree struct {
	tree.Tree[*NamedImage]
}

func NewBuildTree() *BuildTree {
	return &BuildTree{make(tree.Tree[*NamedImage])}
}

func (t *BuildTree) Insert(image *NamedImage) error {
	from := image.From.String()

	for _, tag := range image.Tags {
		ref, err := reference.WithTag(image.Named, tag)
		if err != nil {
			return err
		}

		if parent := t.Tree.Insert(from, ref.String(), image).Parent; parent.Value == nil {
			parent.Value = &NamedImage{
				Named: image.From,
				Tags:  []string{image.From.Tag()},
			}
		}
	}

	return nil
}

func (t *BuildTree) Walk(walker tree.Walker[*NamedImage]) error {
	visited := make(map[*NamedImage]struct{})
	return t.AsNode().Walk(func(level int, name string, node *tree.Node[*NamedImage]) error {
		if _, ok := visited[node.Value]; ok {
			return nil
		} else {
			visited[node.Value] = struct{}{}
		}

		return walker(level, name, node)
	})
}

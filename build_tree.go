package clade

import (
	"fmt"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/tree"
)

type BuildTree struct {
	tree.Tree[*ResolvedImage]
	TagsByName map[string][]string
}

func NewBuildTree() *BuildTree {
	return &BuildTree{
		Tree:       make(tree.Tree[*ResolvedImage]),
		TagsByName: make(map[string][]string),
	}
}

func (t *BuildTree) Insert(image *ResolvedImage) error {
	name := image.Name()
	t.TagsByName[name] = append(t.TagsByName[name], image.Tags...)

	from := image.From.String()

	for _, tag := range image.Tags {
		ref, err := reference.WithTag(image.Named, tag)
		if err != nil {
			return fmt.Errorf("%w: %s", err, tag)
		}

		if parent := t.Tree.Insert(from, ref.String(), image).Parent; parent.Value == nil {
			parent.Value = &ResolvedImage{
				Named: image.From,
				Tags:  []string{image.From.Tag()},
			}
		}
	}

	return nil
}

func (t *BuildTree) Walk(walker tree.Walker[*ResolvedImage]) error {
	visited := make(map[*ResolvedImage]struct{})
	return t.AsNode().Walk(func(level int, name string, node *tree.Node[*ResolvedImage]) error {
		if _, ok := visited[node.Value]; ok {
			return nil
		} else {
			visited[node.Value] = struct{}{}
		}

		return walker(level, name, node)
	})
}

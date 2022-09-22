package internal

import (
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
)

type ExpandTree struct {
	tree.Tree[[]*clade.NamedImage]
}

func NewExpandTree() *ExpandTree {
	return &ExpandTree{make(tree.Tree[[]*clade.NamedImage])}
}

func (t *ExpandTree) Insert(image *clade.NamedImage) error {
	var images []*clade.NamedImage

	node, ok := t.Tree[image.Name()]
	if ok {
		images = append(node.Value, image)
	} else {
		images = []*clade.NamedImage{image}
	}

	t.Tree.Insert(image.From.Name(), image.Name(), images)
	return nil
}

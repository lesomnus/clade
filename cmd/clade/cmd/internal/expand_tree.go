package internal

import (
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
)

type ExpandTree struct {
	tree.Tree[*clade.NamedImage]
}

func NewExpandTree() *ExpandTree {
	return &ExpandTree{make(tree.Tree[*clade.NamedImage])}
}

func (t *ExpandTree) Insert(image *clade.NamedImage) error {
	t.Tree.Insert(image.From.Name(), image.Name(), image)
	return nil
}

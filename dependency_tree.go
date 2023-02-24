package clade

import (
	"github.com/lesomnus/clade/tree"
)

type DependencyTree struct {
	tree.Tree[[]*Image]
}

func NewDependencyTree() *DependencyTree {
	return &DependencyTree{make(tree.Tree[[]*Image])}
}

func (t *DependencyTree) Insert(image *Image) {
	var images []*Image

	node, ok := t.Tree[image.Name()]
	if ok {
		images = append(node.Value, image)
	} else {
		images = []*Image{image}
	}

	t.Tree.Insert(image.From.Primary.Name(), image.Name(), images)
}

package clade

import "github.com/lesomnus/clade/tree"

type DependencyTree struct {
	tree.Tree[[]*NamedImage]
}

func NewDependencyTree() *DependencyTree {
	return &DependencyTree{make(tree.Tree[[]*NamedImage])}
}

func (t *DependencyTree) Insert(image *NamedImage) {
	var images []*NamedImage

	node, ok := t.Tree[image.Name()]
	if ok {
		images = append(node.Value, image)
	} else {
		images = []*NamedImage{image}
	}

	t.Tree.Insert(image.From.Name(), image.Name(), images)
}

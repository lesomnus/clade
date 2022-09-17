package tree

type Node[V any] struct {
	Parent   *Node[V]
	Children map[string]*Node[V]

	Value V
}

type Tree[V any] map[string]*Node[V]

func (t Tree[V]) AsNode() *Node[V] {
	children := make(map[string]*Node[V])
	for name, node := range t {
		if node.Parent != nil {
			continue
		}

		children[name] = node
	}

	return &Node[V]{
		Parent:   nil,
		Children: children,
	}
}

func (t Tree[V]) Insert(pname string, name string, value V) *Node[V] {
	parent, ok := t[pname]
	if !ok {
		parent = &Node[V]{
			Parent:   nil,
			Children: make(map[string]*Node[V]),
		}

		t[pname] = parent
	}

	node, ok := t[name]
	if !ok {
		node = &Node[V]{
			Parent:   parent,
			Children: make(map[string]*Node[V]),
			Value:    value,
		}

		t[name] = node
	} else {
		node.Parent = parent
		node.Value = value
	}

	parent.Children[name] = node

	return node
}

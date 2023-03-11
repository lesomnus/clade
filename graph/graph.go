package graph

type Node[V any] struct {
	key string

	Value V
	Prev  map[string]*Node[V] // Predecessors
	Next  map[string]*Node[V] // Successors
}

func (n *Node[V]) Key() string {
	return n.key
}

func (n *Node[V]) ConnectTo(target *Node[V]) {
	n.Next[target.key] = target
	target.Prev[n.key] = n
}

type Graph[V any] struct {
	entries map[string]*Node[V]
}

func NewGraph[V any]() *Graph[V] {
	return &Graph[V]{
		entries: make(map[string]*Node[V]),
	}
}

func (g *Graph[V]) Get(key string) (*Node[V], bool) {
	node, ok := g.entries[key]
	return node, ok
}

func (g *Graph[V]) Put(key string, value V) *Node[V] {
	node := &Node[V]{
		key: key,

		Value: value,
		Prev:  make(map[string]*Node[V]),
		Next:  make(map[string]*Node[V]),
	}

	g.entries[key] = node
	return node
}

func (g *Graph[V]) GetOrPut(key string, value V) *Node[V] {
	node, ok := g.Get(key)
	if !ok {
		node = g.Put(key, value)
	}

	return node
}

func (g *Graph[V]) Roots() []*Node[V] {
	rst := make([]*Node[V], 0)
	for _, node := range g.entries {
		if len(node.Prev) > 0 {
			continue
		}

		rst = append(rst, node)
	}

	return rst
}

func (g *Graph[V]) PseudoRoot() *Node[V] {
	next := make(map[string]*Node[V])
	for key, node := range g.entries {
		if len(node.Prev) > 0 {
			continue
		}

		next[key] = node
	}

	return &Node[V]{
		Next: next,
	}
}

func (g *Graph[V]) Leaves() []*Node[V] {
	rst := make([]*Node[V], 0)
	for _, node := range g.entries {
		if len(node.Next) > 0 {
			continue
		}

		rst = append(rst, node)
	}

	return rst
}

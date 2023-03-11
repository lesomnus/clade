package graph

type SnapshotEntry struct {
	Level uint
	Group map[uint]bool
}

type Snapshot map[string]*SnapshotEntry

func (s Snapshot) get(key string) *SnapshotEntry {
	entry, ok := s[key]
	if !ok {
		entry = &SnapshotEntry{
			Group: make(map[uint]bool),
		}
		s[key] = entry
	}

	return entry
}

func (g *Graph[V]) Snapshot(keyer func(*Node[V]) string) Snapshot {
	if keyer == nil {
		keyer = func(node *Node[V]) string { return node.key }
	}

	rst := make(Snapshot)

	var visit func(level uint, node *Node[V])
	visit = func(level uint, node *Node[V]) {
		entry := rst.get(keyer(node))
		if entry.Level < level {
			entry.Level = level
		}

		level = level + 1
		for _, next := range node.Next {
			visit(level, next)
		}
	}
	for _, node := range g.Roots() {
		visit(0, node)
	}

	visit = func(group uint, node *Node[V]) {
		entry := rst.get(keyer(node))
		entry.Group[group] = true

		for _, prev := range node.Prev {
			visit(group, prev)
		}
	}
	group := uint(0)
	for _, node := range g.Leaves() {
		visit(group, node)
		group = group + 1
	}

	return rst
}

package tree

import "errors"

type Walker[V any] func(level int, name string, node *Node[V]) error

var (
	WalkContinue = errors.New("") //lint:ignore ST1012 Special error for indicating
	WalkBreak    = errors.New("") //lint:ignore ST1012 Special error for indicating
)

func walk[V any](node *Node[V], level int, walker Walker[V]) error {
	for name, child := range node.Children {
		if err := walker(level, name, child); err != nil {
			if errors.Is(err, WalkContinue) {
				continue
			} else if errors.Is(err, WalkBreak) {
				return nil
			} else {
				return err
			}
		}

		if err := walk(child, level+1, walker); err != nil {
			return err
		}
	}

	return nil
}

func Walk[V any](node *Node[V], walker Walker[V]) error {
	return walk(node, 0, walker)
}

func (n *Node[V]) Walk(walker Walker[V]) error {
	return Walk(n, walker)
}

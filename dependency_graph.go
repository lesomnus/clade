package clade

import "github.com/lesomnus/clade/graph"

type DependencyGraph struct {
	*graph.Graph[[]*Image]
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		Graph: graph.NewGraph[[]*Image](),
	}
}

func (g *DependencyGraph) Put(image *Image) *graph.Node[[]*Image] {
	node := g.Graph.GetOrPut(image.Name(), []*Image{})
	node.Value = append(node.Value, image)

	prev_names := make(map[string]bool, 1+len(image.From.Secondaries))
	prev_names[image.From.Primary.Name()] = true
	for _, base_image := range image.From.Secondaries {
		prev_names[base_image.Name()] = true
	}

	for prev_name := range prev_names {
		g.Graph.GetOrPut(prev_name, []*Image{}).
			ConnectTo(node)
	}

	return node
}

func (g *DependencyGraph) Snapshot() graph.Snapshot {
	return g.Graph.Snapshot(nil)
}

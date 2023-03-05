package clade

import (
	"fmt"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/graph"
	"golang.org/x/exp/slices"
)

type BuildGraph struct {
	*graph.Graph[*ResolvedImage]
	tags_by_name map[string][]string
}

func NewBuildGraph() *BuildGraph {
	return &BuildGraph{
		Graph:        graph.NewGraph[*ResolvedImage](),
		tags_by_name: make(map[string][]string),
	}
}

func (g *BuildGraph) TagsByName(named reference.Named) ([]string, bool) {
	tags, ok := g.tags_by_name[named.Name()]
	if !ok {
		return nil, false
	}

	return slices.Clone(tags), true
}

func (g *BuildGraph) Put(image *ResolvedImage) ([]*graph.Node[*ResolvedImage], error) {
	if len(image.Tags) == 0 {
		return nil, fmt.Errorf("no tags")
	}

	// Put current references.
	refs := make(map[string]reference.NamedTagged, len(image.Tags))
	for _, tag := range image.Tags {
		tagged, err := reference.WithTag(image.Named, tag)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", tag, err)
		}

		refs[tagged.String()] = tagged
	}

	nodes := make([]*graph.Node[*ResolvedImage], 0, len(refs))
	for _, ref := range refs {
		node, ok := g.Graph.Get(ref.String())
		if !ok {
			node = g.Graph.Put(ref.String(), image)
		} else if node.Value == nil {
			node.Value = image
		} else if node.Value != image {
			return nil, fmt.Errorf("duplicate: %s", ref.String())
		}

		nodes = append(nodes, node)
	}

	tags_by_name, ok := g.tags_by_name[image.Name()]
	if !ok {
		tags_by_name = make([]string, 0, len(nodes))
	}
	for _, ref := range refs {
		tags_by_name = append(tags_by_name, ref.Tag())
	}
	g.tags_by_name[image.Name()] = tags_by_name

	// Put previous references.
	refs_prev := make(map[string]bool, 1+len(image.From.Secondaries))
	for _, ref := range image.From.All() {
		refs_prev[ref.String()] = true
	}

	for key := range refs_prev {
		node_prev := g.Graph.GetOrPut(key, nil)
		for _, node_curr := range nodes {
			node_prev.ConnectTo(node_curr)
		}
	}

	return nodes, nil
}

func (g *BuildGraph) Snapshot() graph.Snapshot {
	snapshot := g.Graph.Snapshot(func(node *graph.Node[*ResolvedImage]) string {
		if node.Value == nil {
			return node.Key()
		} else {
			tagged, err := node.Value.Tagged()
			if err != nil {
				panic(fmt.Sprintf("tag must be valid: %s", err.Error()))
			}
			return tagged.String()
		}
	})

	return snapshot
}

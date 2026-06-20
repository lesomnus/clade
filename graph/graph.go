// Package graph builds the dependency graph of build target images from a set
// of ports and marks which targets are outdated with respect to their base.
//
// Each port is expanded into one or more concrete target images by selecting
// upstream versions and rendering the build tag template. A container port
// whose source.repo is the build.repo of another port forms an internal edge;
// such ports are expanded after their upstream so the upstream's produced tags
// are available.
package graph

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/lesomnus/clade/compare"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
	"github.com/lesomnus/clade/registry"
	"github.com/lesomnus/clade/source"
	"github.com/lesomnus/clade/tag"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Builder constructs a graph from ports using a registry for metadata. Each
// port's outdated-comparison chain is built from its own compare config (or the
// default for its source kind).
type Builder struct {
	Registry registry.Registry
}

// Build expands the ports into a graph and marks outdated nodes. The returned
// graph's nodes are ordered topologically (parents before children).
func (b *Builder) Build(ctx context.Context, ports []*port.Port) (*cladev1.Graph, error) {
	ordered, err := topoSort(ports)
	if err != nil {
		return nil, err
	}

	by_repo := map[string]*port.Port{}
	for _, p := range ports {
		by_repo[p.Build.Repo] = p
	}

	expanded := map[string][]string{} // build.repo -> produced target tags
	nodes := []*cladev1.Node{}
	node_by_id := map[string]*cladev1.Node{}
	chains := map[string]compare.Chain{} // port dir -> comparator chain

	for _, p := range ordered {
		var parent_tags []string
		if up, internal := by_repo[p.Source.Repo]; p.Source.Kind == "container" && internal && up != p {
			// Internal edge: reuse the upstream port's produced tags.
			parent_tags = expanded[p.Source.Repo]
		} else {
			src, err := source.New(p.Source.Kind, p.Source.Params, source.Deps{Tags: b.Registry.Tags})
			if err != nil {
				return nil, fmt.Errorf("port %q: %w", p.Dir, err)
			}
			vs, err := src.Versions(ctx)
			if err != nil {
				return nil, fmt.Errorf("list versions for port %q: %w", p.Dir, err)
			}
			parent_tags = vs
		}

		chain, err := compareChain(p)
		if err != nil {
			return nil, fmt.Errorf("port %q: %w", p.Dir, err)
		}
		chains[p.Dir] = chain

		selector, err := tag.New(p.Select.Kind, p.Select.Params)
		if err != nil {
			return nil, fmt.Errorf("port %q: %w", p.Dir, err)
		}
		matched, err := selector.Select(parent_tags)
		if err != nil {
			return nil, fmt.Errorf("select tags for port %q: %w", p.Dir, err)
		}

		tmpls := make([]*template.Template, len(p.Build.Tags))
		for i, t := range p.Build.Tags {
			tmpls[i], err = template.New(p.Dir).Option("missingkey=error").Parse(t)
			if err != nil {
				return nil, fmt.Errorf("parse build tag for port %q: %w", p.Dir, err)
			}
		}

		for _, m := range matched {
			// Render every build tag for this upstream tag. They all point to
			// the same image, so collect their full references.
			var refs, tags []string
			for _, tmpl := range tmpls {
				var sb strings.Builder
				if err := tmpl.Execute(&sb, m.Data); err != nil {
					return nil, fmt.Errorf("render build tag for port %q tag %q: %w", p.Dir, m.Tag, err)
				}
				target_tag := sb.String()
				target_ref := p.Build.Repo + ":" + target_tag
				// matched is ordered newest first, so a reference already taken
				// belongs to a newer image; leave a floating tag (e.g. "1") on it.
				if _, taken := node_by_id[target_ref]; taken {
					continue
				}
				tags = append(tags, target_tag)
				refs = append(refs, target_ref)
			}
			if len(refs) == 0 {
				continue
			}

			// A container source provides the base image; other sources (e.g.
			// http) have no upstream image, so the Dockerfile sets its own FROM.
			base_ref := ""
			if p.Source.Kind == "container" {
				base_ref = p.Source.Repo + ":" + m.Tag
			}
			node := &cladev1.Node{
				Id:      refs[0],
				Tags:    refs,
				Base:    base_ref,
				BaseTag: m.Tag,
				Port:    p.Dir,
				Image:   &cladev1.Image{Repo: p.Build.Repo, Tag: tags[0]},
			}
			if parent, ok := node_by_id[base_ref]; ok {
				node.Parents = []string{parent.Id}
			}

			for i, ref := range refs {
				node_by_id[ref] = node
				expanded[p.Build.Repo] = append(expanded[p.Build.Repo], tags[i])
			}
			nodes = append(nodes, node)
		}
	}

	if err := b.markOutdated(ctx, nodes, node_by_id, chains); err != nil {
		return nil, err
	}
	return &cladev1.Graph{Nodes: nodes}, nil
}

// compareChain builds a port's outdated-comparison chain from its own compare
// config, or the default for its source kind when the port declares none.
func compareChain(p *port.Port) (compare.Chain, error) {
	var specs []compare.Spec
	if len(p.Compare) > 0 {
		specs = make([]compare.Spec, 0, len(p.Compare))
		for _, c := range p.Compare {
			specs = append(specs, compare.Spec{Kind: c.Kind, Params: c.Params})
		}
	} else {
		specs = compare.DefaultFor(p.Source.Kind)
	}
	return compare.NewChain(specs)
}

// markOutdated fetches target/base metadata and sets the outdated flag. Nodes
// are visited in topological order so a parent's flag is final before its
// children are evaluated.
func (b *Builder) markOutdated(ctx context.Context, nodes []*cladev1.Node, node_by_id map[string]*cladev1.Node, chains map[string]compare.Chain) error {
	stats := map[string]*registry.ImageInfo{}
	stat := func(ref string) (*registry.ImageInfo, error) {
		if info, ok := stats[ref]; ok {
			return info, nil
		}
		info, err := b.Registry.Stat(ctx, ref)
		if errors.Is(err, registry.ErrNotExist) {
			stats[ref] = nil
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		stats[ref] = info
		return info, nil
	}

	for _, node := range nodes {
		target_info, err := stat(node.Id)
		if err != nil {
			return fmt.Errorf("stat target %q: %w", node.Id, err)
		}
		if target_info != nil {
			node.Image = imageOf(node.Image.Repo, node.Image.Tag, target_info)
		}

		// A rebuilt base invalidates its descendants.
		if anyParentOutdated(node, node_by_id) {
			node.Outdated = true
			continue
		}

		// A missing target must be built.
		if target_info == nil {
			node.Outdated = true
			continue
		}

		// An empty chain (e.g. an http source) judges by existence only: the
		// primary tag is present, so the target is up to date.
		chain := chains[node.Port]
		if len(chain) == 0 {
			continue
		}

		base_info, err := stat(node.Base)
		if err != nil {
			return fmt.Errorf("stat base %q: %w", node.Base, err)
		}
		if base_info == nil {
			// The base does not exist yet (e.g. an internal parent that is
			// about to be built), so this target is outdated as well.
			node.Outdated = true
			continue
		}

		outdated, err := chain.IsOutdated(ctx, compare.OfImage(base_info), compare.OfImage(target_info))
		if err != nil {
			return fmt.Errorf("compare %q: %w", node.Id, err)
		}
		node.Outdated = outdated
	}
	return nil
}

func anyParentOutdated(node *cladev1.Node, node_by_id map[string]*cladev1.Node) bool {
	for _, pid := range node.Parents {
		if parent, ok := node_by_id[pid]; ok && parent.Outdated {
			return true
		}
	}
	return false
}

func imageOf(repo, tag string, info *registry.ImageInfo) *cladev1.Image {
	return &cladev1.Image{
		Repo:    repo,
		Tag:     tag,
		Digest:  info.Digest,
		Created: timestamppb.New(info.Created),
		Labels:  info.Labels,
	}
}

// topoSort orders ports so that a port whose parent.repo is the build.repo of
// another port comes after that port. It returns an error on a cycle.
func topoSort(ports []*port.Port) ([]*port.Port, error) {
	by_repo := map[string]*port.Port{}
	for _, p := range ports {
		by_repo[p.Build.Repo] = p
	}

	indeg := map[*port.Port]int{}
	children := map[*port.Port][]*port.Port{}
	for _, p := range ports {
		indeg[p] = 0
	}
	for _, p := range ports {
		if up, ok := by_repo[p.Source.Repo]; ok && up != p {
			children[up] = append(children[up], p)
			indeg[p]++
		}
	}

	queue := []*port.Port{}
	for _, p := range ports {
		if indeg[p] == 0 {
			queue = append(queue, p)
		}
	}

	out := make([]*port.Port, 0, len(ports))
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		out = append(out, p)
		for _, c := range children[p] {
			indeg[c]--
			if indeg[c] == 0 {
				queue = append(queue, c)
			}
		}
	}

	if len(out) != len(ports) {
		return nil, fmt.Errorf("dependency cycle among ports")
	}
	return out, nil
}

// Package graph builds the dependency graph of build target images from a set
// of ports and marks which targets are outdated with respect to their base.
//
// Each port is expanded into one or more concrete target images by selecting
// upstream tags and rendering the build tag template. A port whose parent.repo
// is the build.repo of another port forms an internal edge; such ports are
// expanded after their upstream so the upstream's produced tags are available.
package graph

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lesomnus/clade/compare"
	cladev1 "github.com/lesomnus/clade/pb/clade/v1"
	"github.com/lesomnus/clade/port"
	"github.com/lesomnus/clade/registry"
	"github.com/lesomnus/clade/tag"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Builder constructs a graph from ports using a registry for metadata and a
// comparator to decide whether targets are outdated.
type Builder struct {
	Registry   registry.Registry
	Comparator compare.Comparator
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

	for _, p := range ordered {
		var parent_tags []string
		if up, internal := by_repo[p.Parent.Repo]; internal && up != p {
			parent_tags = expanded[p.Parent.Repo]
		} else {
			parent_tags, err = b.Registry.Tags(ctx, p.Parent.Repo)
			if err != nil {
				return nil, fmt.Errorf("list tags for port %q: %w", p.Dir, err)
			}
		}

		selector, err := tag.New(p.Parent.Target.Kind, p.Parent.Target.Params)
		if err != nil {
			return nil, fmt.Errorf("port %q: %w", p.Dir, err)
		}
		matched, err := selector.Select(parent_tags)
		if err != nil {
			return nil, fmt.Errorf("select tags for port %q: %w", p.Dir, err)
		}

		tmpl, err := template.New(p.Dir).Option("missingkey=error").Parse(p.Build.Tag)
		if err != nil {
			return nil, fmt.Errorf("parse build tag for port %q: %w", p.Dir, err)
		}

		for _, m := range matched {
			var sb strings.Builder
			if err := tmpl.Execute(&sb, m.Data); err != nil {
				return nil, fmt.Errorf("render build tag for port %q tag %q: %w", p.Dir, m.Tag, err)
			}
			target_tag := sb.String()
			base_ref := p.Parent.Repo + ":" + m.Tag
			target_ref := p.Build.Repo + ":" + target_tag
			if _, dup := node_by_id[target_ref]; dup {
				continue
			}

			node := &cladev1.Node{
				Id:         target_ref,
				Base:       base_ref,
				Port:       p.Dir,
				Dockerfile: filepath.Join(p.Dir, "Dockerfile"),
				Context:    p.Dir,
				Args:       map[string]string{"BASE": base_ref},
				Image:      &cladev1.Image{Repo: p.Build.Repo, Tag: target_tag},
			}
			if parent, ok := node_by_id[base_ref]; ok {
				node.Parents = []string{parent.Id}
			}

			expanded[p.Build.Repo] = append(expanded[p.Build.Repo], target_tag)
			nodes = append(nodes, node)
			node_by_id[target_ref] = node
		}
	}

	if err := b.markOutdated(ctx, nodes, node_by_id); err != nil {
		return nil, err
	}
	return &cladev1.Graph{Nodes: nodes}, nil
}

// markOutdated fetches target/base metadata and sets the outdated flag. Nodes
// are visited in topological order so a parent's flag is final before its
// children are evaluated.
func (b *Builder) markOutdated(ctx context.Context, nodes []*cladev1.Node, node_by_id map[string]*cladev1.Node) error {
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

		outdated, err := b.Comparator.IsOutdated(ctx, base_info, target_info)
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
		if up, ok := by_repo[p.Parent.Repo]; ok && up != p {
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

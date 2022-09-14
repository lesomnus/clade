package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blang/semver/v4"
	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
	"gopkg.in/yaml.v3"
)

// TODO: move expanders into lesomnus/clade

type BuildTree struct {
	tree.Tree[*clade.NamedImage]
}

func NewBuildTree() *BuildTree {
	return &BuildTree{make(tree.Tree[*clade.NamedImage])}
}

func (t *BuildTree) Insert(image *clade.NamedImage) error {
	from := image.From.String()
	for _, tag := range image.Tags {
		ref, err := reference.WithTag(image.Named, tag)
		if err != nil {
			return err
		}

		if parent := t.Tree.Insert(from, ref.String(), image).Parent; parent.Value == nil {
			parent.Value = &clade.NamedImage{
				Named: image.From,
				Tags:  []string{image.From.Tag()},
			}
		}
	}

	return nil
}

func (t *BuildTree) Walk(walker tree.Walker[*clade.NamedImage]) error {
	visited := make(map[*clade.NamedImage]struct{})
	return t.AsNode().Walk(func(level int, name string, node *tree.Node[*clade.NamedImage]) error {
		if _, ok := visited[node.Value]; ok {
			return nil
		} else {
			visited[node.Value] = struct{}{}
		}

		return walker(level, name, node)
	})
}

// TODO: make general plan for expanding
func ExpandImage(ctx context.Context, image *clade.NamedImage, bt *BuildTree) ([]*clade.NamedImage, error) {
	switch image.From.(type) {
	case clade.RefNamedPipelineTagged:
		return ExpandByPipeline(ctx, image, bt)
	case clade.RefNamedRegexTagged:
		return ExpandByRegex(ctx, image)
	default:
		return []*clade.NamedImage{image}, nil
	}
}

func ExpandByRegex(ctx context.Context, image *clade.NamedImage) ([]*clade.NamedImage, error) {
	repo, err := NewRepository(image.From)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	tags, err := repo.Tags(ctx).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	images, err := clade.NewRegexExpander(tags)(ctx, image)
	if err != nil {
		return nil, fmt.Errorf("failed to expand image with regex: %w", err)
	}

	for i := 0; i < len(images); i++ {
		for j := i + 1; j < len(images); j++ {
			clade.DeduplicateBySemver(&images[i].Tags, &images[j].Tags)
		}
	}

	return images, nil
}

func ExpandByPipeline(ctx context.Context, image *clade.NamedImage, bt *BuildTree) ([]*clade.NamedImage, error) {
	tagged, ok := image.From.(clade.RefNamedPipelineTagged)
	if !ok {
		return nil, errors.New("not a pipeline tagged")
	}

	tags := make([]string, 0)
	for _, node := range bt.Tree {
		if node.Value.Name() != tagged.Name() {
			continue
		}

		tags = append(tags, node.Value.Tags...)
	}

	sb := strings.Builder{}
	if tmpl, err := template.New("").
		Funcs(template.FuncMap{
			"localTags":    func() []string { return tags },
			"semverLatest": clade.SemverStringLatest,
		}).
		Parse(tagged.PipelineExpr()); err != nil {
		return nil, fmt.Errorf("failed to parse pipeline: %w", err)
	} else if err := tmpl.Execute(&sb, nil); err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	tag := sb.String()

	// TODO: tag may not be a semver string in the future.

	// Find existing tag.
	// Tag is resolved to full valid semver notation e.g. 1.0.0,
	// but there may be not exists but 1.0 or 1.
	if v, err := semver.ParseTolerant(tag); err != nil {
		return nil, fmt.Errorf("invalid semver %s: %w", tag, err)
	} else {
		for _, t := range tags {
			if sv, err := semver.ParseTolerant(t); err != nil {
				continue
			} else if v.EQ(sv) {
				tag = t
				break
			}
		}
	}

	from, err := reference.WithTag(image.From, tag)
	if err != nil {
		return nil, fmt.Errorf("invalid tag %s: %w", tag, err)
	}

	rst := &clade.NamedImage{}
	*rst = *image
	rst.Tags = []string{tag}
	rst.From = from

	return []*clade.NamedImage{rst}, nil
}

// TODO: move to lesomnus/clade?
func ReadPort(path string) (*clade.Port, error) {
	port_def_path := filepath.Join(path, "port.yaml")

	data, err := os.ReadFile(port_def_path)
	if err != nil {
		return nil, err
	}

	port := &clade.Port{}
	if err := yaml.Unmarshal(data, &port); err != nil {
		return nil, fmt.Errorf("failed to unmarshal port at %s: %w", port_def_path, err)
	}

	if p, err := ResolvePath(path, port.Dockerfile, "Dockerfile"); err != nil {
		return nil, fmt.Errorf("failed to resolve path to Dockerfile: %w", err)
	} else {
		port.Dockerfile = p
	}

	if p, err := ResolvePath(path, port.ContextPath, "."); err != nil {
		return nil, fmt.Errorf("failed to resolve path to context: %w", err)
	} else {
		port.ContextPath = p
	}

	for i, image := range port.Images {
		if p, err := ResolvePath(path, image.Dockerfile, port.Dockerfile); err != nil {
			return nil, fmt.Errorf("failed to resolve path to Dockerfile: %w", err)
		} else {
			port.Images[i].Dockerfile = p
		}

		if p, err := ResolvePath(path, image.ContextPath, port.ContextPath); err != nil {
			return nil, fmt.Errorf("failed to resolve path to context: %w", err)
		} else {
			port.Images[i].ContextPath = p
		}
	}

	return port, nil
}

func ReadPorts(path string) ([]*clade.Port, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	ports := make([]*clade.Port, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		port_path := filepath.Join(path, entry.Name())
		port, err := ReadPort(port_path)
		if err != nil {
			return nil, fmt.Errorf("failed to read port at %s: %w", port_path, err)
		}

		ports = append(ports, port)
	}

	return ports, nil
}

func expandPriorityOf(image *clade.NamedImage) int {
	switch image.From.(type) {
	case clade.RefNamedPipelineTagged:
		return 2
	case clade.RefNamedRegexTagged:
		return 1
	default:
		return 0
	}
}

func lowestExpandPriorityOf(images []*clade.NamedImage) int {
	rst := 0
	for _, image := range images {
		p := expandPriorityOf(image)
		if rst < p {
			rst = p
		}
	}

	return rst
}

func LoadBuildTreeFromPorts(ctx context.Context, bt *BuildTree, path string) error {
	ports, err := ReadPorts(path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	et := NewExpandTree()
	for _, port := range ports {
		images, err := port.ParseImages()
		if err != nil {
			return fmt.Errorf("failed to parse image from port %s: %w", port.Name, err)
		}

		for _, image := range images {
			if err := et.Insert(image); err != nil {
				return fmt.Errorf("failed to insert image %s into expand tree: %w", image.String(), err)
			}
		}
	}

	et.AsNode().Walk(func(level int, name string, node *tree.Node[*clade.NamedImage]) error {
		if level == 0 {
			return nil
		}

		images, err := ExpandImage(ctx, node.Value, bt)
		if err != nil {
			return fmt.Errorf("failed to expand image %s: %w", node.Value.String(), err)
		}

		for _, image := range images {
			if err := bt.Insert(image); err != nil {
				return fmt.Errorf("failed to insert image %s into build tree: %w", image.String(), err)
			}
		}

		return nil
	})

	return nil
}

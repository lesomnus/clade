package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/pipeline"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/sv"
	"github.com/lesomnus/clade/tree"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// TODO: move expanders into lesomnus/clade?

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

func ExpandImage(ctx context.Context, image *clade.NamedImage, bt *BuildTree) ([]*clade.NamedImage, error) {
	switch image.From.(type) {
	case clade.RefNamedPipelineTagged:
		return ExpandByPipeline(ctx, image, bt)
	default:
		return []*clade.NamedImage{image}, nil
	}
}

func ExpandByPipeline(ctx context.Context, image *clade.NamedImage, bt *BuildTree) ([]*clade.NamedImage, error) {
	tagged, ok := image.From.(clade.RefNamedPipelineTagged)
	if !ok {
		return nil, errors.New("not a pipeline tagged")
	}

	local_tags := make([]string, 0)
	for _, node := range bt.Tree {
		if node.Value.Name() != tagged.Name() {
			continue
		}

		local_tags = append(local_tags, node.Value.Tags...)
	}

	remote_tags := make([]string, 0)
	if tagged.Pipeline().HasFunction("remoteTags") {
		repo, err := NewRepository(image.From)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		tags, err := repo.Tags(ctx).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}

		remote_tags = tags
	}

	exe := pipeline.Executor{
		Funcs: plf.FuncMap(),
	}

	maps.Copy(exe.Funcs, pipeline.FuncMap{
		"localTags":  func() []string { return local_tags },
		"remoteTags": func() []string { return remote_tags },
	})

	rst, err := exe.Execute(tagged.Pipeline())
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	var versions []sv.Version
	if v, ok := rst.(sv.Version); ok {
		versions = []sv.Version{v}
	} else if rst_t := reflect.TypeOf(rst); rst_t.Kind() == reflect.Slice {
		vs := reflect.ValueOf(rst)
		versions = make([]sv.Version, vs.Len())
		for i := range versions {
			v, ok := vs.Index(i).Interface().(sv.Version)
			if !ok {
				panic("currently only semver.Version is supported")
			}

			versions[i] = v
		}
	} else {
		panic("currently only server.Version is supported")
	}

	images := make([]*clade.NamedImage, 0, len(versions))
	for _, version := range versions {
		tags := slices.Clone(image.Tags)
		for i, tag := range tags {
			tmpl, err := template.New("").Parse(tag)
			if err != nil {
				continue
			}

			var sb strings.Builder
			tmpl.Execute(&sb, version)
			tags[i] = sb.String()
		}

		tag := version.String()

		from, err := reference.WithTag(image.From, tag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag %s: %w", tag, err)
		}

		img := &clade.NamedImage{}
		*img = *image
		img.Tags = tags
		img.From = from

		images = append(images, img)
	}

	// TODO: make it as functions.
	for i := 0; i < len(images); i++ {
		for j := i + 1; j < len(images); j++ {
			clade.DeduplicateBySemver(&images[i].Tags, &images[j].Tags)
		}
	}

	return images, nil
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

	return et.AsNode().Walk(func(level int, name string, node *tree.Node[[]*clade.NamedImage]) error {
		if level == 0 {
			return nil
		}

		for _, image := range node.Value {
			images, err := ExpandImage(ctx, image, bt)
			if err != nil {
				return fmt.Errorf("failed to expand image %s: %w", image.String(), err)
			}

			for _, image := range images {
				if len(image.Tags) == 0 {
					continue
				}

				if err := bt.Insert(image); err != nil {
					return fmt.Errorf("failed to insert image %s into build tree: %w", image.String(), err)
				}
			}
		}

		return nil
	})
}

package internal

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/pipeline"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/sv"
	"github.com/lesomnus/clade/tree"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func ExpandImage(ctx context.Context, image *clade.Image, bt *clade.BuildTree) ([]*clade.Image, error) {
	local_tags := make([]string, 0)
	if image.From.Pipeline().HasFunction("localTags") {
		for _, node := range bt.Tree {
			if node.Value.Name() != image.From.Name() {
				continue
			}

			local_tags = append(local_tags, node.Value.Tags...)
		}
	}

	remote_tags := make([]string, 0)
	if image.From.Pipeline().HasFunction("remoteTags") {
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

	rst, err := exe.Execute(image.From.Pipeline())
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	versions := make([]sv.Version, len(rst))
	for i := range versions {
		v, ok := reflect.ValueOf(rst[i]).Interface().(sv.Version)
		if !ok {
			panic("currently only semver.Version is supported")
		}

		versions[i] = v
	}

	images := make([]*clade.Image, 0, len(versions))
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

		from, err := clade.RefWithTag(image.From, tag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag %s: %w", tag, err)
		}

		img := &clade.Image{}
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

func LoadBuildTreeFromPorts(ctx context.Context, bt *clade.BuildTree, path string) error {
	ports, err := ReadPorts(path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	dt := clade.NewDependencyTree()
	for _, port := range ports {
		for _, image := range port.Images {
			dt.Insert(image)
		}
	}

	return dt.AsNode().Walk(func(level int, name string, node *tree.Node[[]*clade.Image]) error {
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

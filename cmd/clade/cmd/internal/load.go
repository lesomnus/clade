package internal

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/tree"
	"github.com/lesomnus/pl"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func hasFunc(pipeline *pl.Pl, name string) bool {
	for _, fn := range pipeline.Funcs {
		if fn.Name == name {
			return true
		}

		for _, arg := range fn.Args {
			if arg.Nested == nil {
				continue
			}

			if hasFunc(arg.Nested, name) {
				return true
			}
		}
	}

	return false
}

func ExpandImage(ctx context.Context, image *clade.Image, bt *clade.BuildTree) ([]*clade.Image, error) {
	executor := pl.NewExecutor()
	maps.Copy(executor.Funcs, plf.Funcs())
	executor.Funcs["tags"] = func() ([]string, error) {
		if tags := bt.TagsByName[image.From.Name()]; len(tags) > 0 {
			return tags, nil
		}

		repo, err := NewRepository(image.From)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		tags, err := repo.Tags(ctx).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}

		return tags, nil
	}

	results, err := executor.Execute(image.From.Pipeline())
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	base_tags := make([]string, len(results))
	for i, result := range results {
		if s, ok := result.(interface{ String() string }); ok {
			base_tags[i] = s.String()
			continue
		}

		v := reflect.ValueOf(result)
		if v.Kind() == reflect.String {
			base_tags[i] = v.String()
			continue
		}

		return nil, fmt.Errorf("failed to resolve string from pipeline result: %v", result)
	}

	images := make([]*clade.Image, 0, len(results))
	for i, result := range results {
		tags := slices.Clone(image.Tags)
		for i, tag := range tags {
			tmpl, err := template.New("").Parse(tag)
			if err != nil {
				continue
			}

			var sb strings.Builder
			tmpl.Execute(&sb, result)
			tags[i] = sb.String()
		}

		tag := base_tags[i]

		tagged, err := reference.WithTag(image.From, tag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag %s: %w", tag, err)
		}

		img := &clade.Image{}
		*img = *image
		img.Tags = tags
		img.From = clade.AsRefNamedPipelineTagged(tagged)

		images = append(images, img)
	}

	// TODO: make it as functions.
	for i := 0; i < len(images); i++ {
		for j := i + 1; j < len(images); j++ {
			clade.DeduplicateBySemver(&images[i].Tags, &images[j].Tags)
		}
	}

	// Test if duplicated tags are exist.
	for i := 0; i < len(images); i++ {
		for j := i + 1; j < len(images); j++ {
			for _, lhs := range images[i].Tags {
				for _, rhs := range images[j].Tags {
					if lhs == rhs {
						return nil, fmt.Errorf("%s: tag is duplicated: %s", image.From.String(), lhs)
					}
				}
			}
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

package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/tree"
	"github.com/lesomnus/pl"
	"golang.org/x/exp/maps"
)

func ExpandImage(ctx context.Context, image *clade.Image, bt *clade.BuildTree, modifiers ...ClientOptionModifier) ([]*clade.ResolvedImage, error) {
	Log.Debug().Str("name", image.Name()).Msg("expand")

	executor := pl.NewExecutor()
	maps.Copy(executor.Funcs, plf.Funcs())
	executor.Convs.MergeWith(plf.Convs())
	executor.Funcs["log"] = func(vs ...string) []string {
		Log.Debug().Str("name", image.Name()).Msg(strings.Join(vs, ", "))
		return vs
	}
	executor.Funcs["tags"] = func() ([]string, error) {
		if tags := bt.TagsByName[image.From.Name()]; len(tags) > 0 {
			return tags, nil
		}

		repo, err := NewRepository(image.From, modifiers...)
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %w", err)
		}

		tags, err := repo.Tags(ctx).All(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags: %w", err)
		}

		return tags, nil
	}

	from_results, err := executor.Execute(image.From.Pipeline(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	delete(executor.Funcs, "tags")

	resolved_images := make([]*clade.ResolvedImage, len(from_results))

	for i, result := range from_results {
		tag := ""
		switch v := result.(type) {
		case interface{ String() string }:
			tag = v.String()
		case string:
			tag = v

		default:
			return nil, fmt.Errorf("failed convert to string from result of base image's tag: %v", v)
		}

		tagged, err := reference.WithTag(image.From, tag)
		if err != nil {
			return nil, fmt.Errorf("invalid tag of base image: %w", err)
		}

		resolved_images[i] = &clade.ResolvedImage{From: tagged}
	}

	for i, result := range from_results {
		tags := make([]string, len(image.Tags))
		for i, tag := range image.Tags {
			tag_results, err := executor.Execute(tag.Pipeline(), result)
			if err != nil {
				return nil, fmt.Errorf("failed to execute pipeline: %w", err)
			}
			if len(tag_results) != 1 {
				return nil, fmt.Errorf("result of pipeline for tags must be size 1 but was %d", len(tag_results))
			}

			v, ok := tag_results[0].(string)
			if !ok {
				return nil, fmt.Errorf("result of pipeline must be string")
			}

			tags[i] = v
		}

		resolved_images[i].Tags = tags
	}

	for i := range resolved_images {
		resolved_images[i].Named = image.Named
		resolved_images[i].Args = image.Args
		resolved_images[i].Dockerfile = image.Dockerfile
		resolved_images[i].ContextPath = image.ContextPath
		resolved_images[i].Platform = image.Platform
	}

	// TODO: make it as functions.
	for i := 0; i < len(resolved_images); i++ {
		for j := i + 1; j < len(resolved_images); j++ {
			clade.DeduplicateBySemver(&resolved_images[i].Tags, &resolved_images[j].Tags)
		}
	}

	// Test if duplicated tags are exist.
	for i := 0; i < len(resolved_images); i++ {
		for j := i + 1; j < len(resolved_images); j++ {
			for _, lhs := range resolved_images[i].Tags {
				for _, rhs := range resolved_images[j].Tags {
					if lhs == rhs {
						return nil, fmt.Errorf("%s: tag is duplicated: %s", image.From.String(), lhs)
					}
				}
			}
		}
	}

	return resolved_images, nil
}

func LoadBuildTreeFromPorts(ctx context.Context, bt *clade.BuildTree, path string, modifiers ...ClientOptionModifier) error {
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
			images, err := ExpandImage(ctx, image, bt, modifiers...)
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

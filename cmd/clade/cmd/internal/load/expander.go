package load

import (
	"context"
	"fmt"
	"strings"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/pl"
	"github.com/rs/zerolog"
	"golang.org/x/exp/maps"
)

type Expander struct {
	Registry *client.Registry
}

func (e *Expander) remoteTags(ctx context.Context, ref reference.Named) ([]string, error) {
	repo, err := e.Registry.Repository(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	tags, err := repo.Tags(ctx).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return tags, nil
}

func (e *Expander) Expand(ctx context.Context, image *clade.Image, bt *clade.BuildTree) ([]*clade.ResolvedImage, error) {
	l := zerolog.Ctx(ctx)
	l.Debug().Str("name", image.Name()).Msg("expand")

	executor := pl.NewExecutor()
	maps.Copy(executor.Funcs, plf.Funcs())
	executor.Convs.MergeWith(plf.Convs())
	executor.Funcs["log"] = func(vs ...string) []string {
		l.Info().Str("name", image.Name()).Msg(strings.Join(vs, ", "))
		return vs
	}
	executor.Funcs["tags"] = func() ([]string, error) {
		if tags := bt.TagsByName[image.From.Name()]; len(tags) > 0 {
			return tags, nil
		}

		return e.remoteTags(ctx, image.From)
	}
	executor.Funcs["tagsOf"] = func(ref string) ([]string, error) {
		if tags := bt.TagsByName[ref]; len(tags) > 0 {
			return tags, nil
		}

		named, err := reference.ParseNamed(ref)
		if err != nil {
			return nil, fmt.Errorf("reference must be fully named: %w", err)
		}

		return e.remoteTags(ctx, named)
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

		resolved_images[i] = &clade.ResolvedImage{
			From: tagged,
			Skip: false,
		}
	}

	invoke := func(pl *pl.Pl, data any) (string, error) {
		tag_results, err := executor.Execute(pl, data)
		if err != nil {
			return "", fmt.Errorf("failed to execute pipeline: %w", err)
		}
		if len(tag_results) != 1 {
			return "", fmt.Errorf("result of pipeline must be size 1 but was %d", len(tag_results))
		}

		v, ok := tag_results[0].(string)
		if !ok {
			return "", fmt.Errorf("result of pipeline must be string")
		}

		return v, nil
	}

	for i, result := range from_results {
		tags := make([]string, len(image.Tags))
		for j, tag := range image.Tags {
			v, err := invoke(tag.Pipeline(), result)
			if err != nil {
				return nil, fmt.Errorf("tags[%d] %s: %w", j, tag.String(), err)
			}

			tags[j] = v
		}

		resolved_images[i].Tags = tags
	}

	for i, result := range from_results {
		args := make(map[string]string)
		for key, arg := range image.Args {
			v, err := invoke(arg.Pipeline(), result)
			if err != nil {
				return nil, fmt.Errorf("args[%s] %s: %w", key, arg.String(), err)
			}

			args[key] = v
		}

		resolved_images[i].Args = args
	}

	for i := range resolved_images {
		resolved_images[i].Named = image.Named
		resolved_images[i].Skip = *image.Skip
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

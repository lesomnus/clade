package clade

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/distribution/reference"
)

type Expander func(ctx context.Context, image *NamedImage) ([]*NamedImage, error)

func NewRegexExpander(tags []string) Expander {
	return func(ctx context.Context, image *NamedImage) ([]*NamedImage, error) {
		expr := image.From.Tag()
		if !(strings.HasPrefix(expr, "/") && strings.HasSuffix(expr, "/")) {
			return []*NamedImage{image}, nil
		}

		pattern, err := regexp.Compile(expr[1 : len(expr)-1])
		if err != nil {
			return nil, fmt.Errorf("failed to compile regex: %w", err)
		}

		rst := make([]*NamedImage, 0)
		for _, tag := range tags {
			match := pattern.FindStringSubmatch(tag)
			if len(match) == 0 {
				continue
			}

			expanded_image := new(NamedImage)
			*expanded_image = *image
			expanded_image.Tags = make([]string, len(image.Tags))
			expanded_image.From, err = reference.WithTag(image.From, tag)
			if err != nil {
				return nil, fmt.Errorf("failed to tag %s: %w", tag, err)
			}
			for i, template := range image.Tags {
				expanded_tag := []byte{}
				for _, submatches := range pattern.FindAllStringSubmatchIndex(tag, -1) {
					expanded_tag = pattern.ExpandString(expanded_tag, template, tag, submatches)
				}

				expanded_image.Tags[i] = string(expanded_tag)
			}

			rst = append(rst, expanded_image)
		}

		return rst, nil
	}
}

package source

import (
	"context"
	"fmt"

	"github.com/goccy/go-yaml"
)

func init() {
	Register("container", newContainer)
}

// containerConfig is the config for the container source.
//
//	source:
//	  kind: container
//	  repo: docker.io/library/golang
type containerConfig struct {
	Repo string `yaml:"repo"`
}

// container lists the tags of an OCI repository as candidate versions.
type container struct {
	repo string
	tags func(ctx context.Context, repo string) ([]string, error)
}

func newContainer(params []byte, deps Deps) (Source, error) {
	cfg := containerConfig{}
	if len(params) > 0 {
		if err := yaml.Unmarshal(params, &cfg); err != nil {
			return nil, fmt.Errorf("decode container source: %w", err)
		}
	}
	if cfg.Repo == "" {
		return nil, fmt.Errorf("container source: repo is required")
	}
	if deps.Tags == nil {
		return nil, fmt.Errorf("container source: tags lister is required")
	}
	return &container{repo: cfg.Repo, tags: deps.Tags}, nil
}

func (c *container) Versions(ctx context.Context) ([]string, error) {
	return c.tags(ctx, c.repo)
}

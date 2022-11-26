package client

import (
	"context"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
)

type distRepository struct {
	distribution.Repository
	named reference.Named
	cache cache.CacheStore
}

func (r *distRepository) Tags(ctx context.Context) distribution.TagService {
	svc := r.Repository.Tags(ctx)

	return &distTagSvc{
		TagService: svc,

		named: r.named,
		cache: r.cache,
	}
}

type distTagSvc struct {
	distribution.TagService
	named reference.Named
	cache cache.CacheStore
}

func (s *distTagSvc) All(ctx context.Context) ([]string, error) {
	if tags, ok := s.cache.GetTags(s.named); ok {
		return tags, nil
	}

	tags, err := s.TagService.All(ctx)
	if err == nil {
		s.cache.SetTags(s.named, tags)
	}

	return tags, err
}

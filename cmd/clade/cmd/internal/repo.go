package internal

import (
	"context"

	"github.com/distribution/distribution/reference"
	"github.com/docker/distribution"
)

type repoWrapper struct {
	distribution.Repository
	ref reference.Named
}

func (r *repoWrapper) Tags(ctx context.Context) distribution.TagService {
	svc := r.Repository.Tags(ctx)

	return tagSvcWrapper{
		TagService: svc,
		ref:        r.ref,
	}
}

type tagSvcWrapper struct {
	distribution.TagService
	ref reference.Named
}

func (s tagSvcWrapper) All(ctx context.Context) ([]string, error) {
	if tags, ok := Cache.GetTags(s.ref); ok {
		return tags, nil
	}

	tags, err := s.TagService.All(ctx)
	if err == nil {
		Cache.SetTags(s.ref, tags)
	}

	return tags, err
}

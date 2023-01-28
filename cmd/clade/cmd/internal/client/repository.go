package client

import (
	"context"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/opencontainers/go-digest"
)

type distRepository struct {
	distribution.Repository
	named reference.Named
	cache cache.CacheStore
}

func (r *distRepository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	svc, err := r.Repository.Manifests(ctx, options...)
	if err != nil {
		return svc, err
	}

	return &manifestSvc{ManifestService: svc, repo: r}, nil
}

func (r *distRepository) Tags(ctx context.Context) distribution.TagService {
	svc := r.Repository.Tags(ctx)

	return &tagSvc{TagService: svc, repo: r}
}

type manifestSvc struct {
	distribution.ManifestService
	repo *distRepository
}

func (s *manifestSvc) Get(ctx context.Context, dgst digest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	c := s.repo.cache.Manifests

	var tagged reference.NamedTagged = nil
	for _, option := range options {
		opt, ok := option.(distribution.WithTagOption)
		if !ok {
			continue
		}

		ref_tagged, err := reference.WithTag(s.repo.named, opt.Tag)
		if err != nil {
			tagged = nil
		} else {
			tagged = ref_tagged
		}
	}

	var (
		manif distribution.Manifest
		ok    bool
	)
	if tagged != nil {
		manif, ok = c.GetByRef(tagged)
	} else {
		manif, ok = c.GetByDigest(dgst)
	}

	if ok {
		return manif, nil
	}

	manif, err := s.ManifestService.Get(ctx, dgst, options...)
	if err != nil {
		return nil, err
	}

	if tagged != nil {
		c.SetByRef(tagged, manif)
	} else {
		c.SetByDigest(dgst, manif)
	}

	return manif, nil
}

type tagSvc struct {
	distribution.TagService
	repo *distRepository
}

func (s *tagSvc) All(ctx context.Context) ([]string, error) {
	if tags, ok := s.repo.cache.Tags.Get(s.repo.named); ok {
		return tags, nil
	}

	tags, err := s.TagService.All(ctx)
	if err == nil {
		s.repo.cache.Tags.Set(s.repo.named, tags)
	}

	return tags, err
}

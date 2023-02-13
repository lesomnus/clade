package cache

import (
	"context"
	"fmt"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/opencontainers/go-digest"
)

type Namespace interface {
	Repository(named reference.Named) (distribution.Repository, error)
}

type remoteRegistry struct {
	cache  *Registry
	remote Namespace
}

func WithRemote(cache *Registry, remote Namespace) Namespace {
	return &remoteRegistry{
		cache:  cache,
		remote: remote,
	}
}

func (r *remoteRegistry) Repository(named reference.Named) (distribution.Repository, error) {
	repo, err := r.remote.Repository(named)
	if err != nil {
		return nil, fmt.Errorf(`create repository "%s": %w`, named.String(), err)
	}

	return &remoteRepository{
		Repository: repo,
		repository: r.cache.repository(named),
	}, nil
}

type remoteRepository struct {
	distribution.Repository
	repository *Repository
}

func (r *remoteRepository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	svc, err := r.Repository.Manifests(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create manifest service: %w", err)
	}

	return &remoteManifestService{
		ManifestService: svc,
		remote:          r,
		cache:           r.repository.manifests(),
	}, nil
}

func (r *remoteRepository) Tags(ctx context.Context) distribution.TagService {
	return &remoteTagService{
		TagService: r.Repository.Tags(ctx),
		cache:      r.repository.tags(),
	}
}

type remoteManifestService struct {
	distribution.ManifestService
	remote *remoteRepository
	cache  *ManifestService
}

func (s *remoteManifestService) Exists(ctx context.Context, dgst digest.Digest) (bool, error) {
	ok, err := s.cache.Exists(ctx, dgst)
	if err == nil {
		return ok, nil
	}

	return s.ManifestService.Exists(ctx, dgst)
}

func (s *remoteManifestService) Get(ctx context.Context, dgst digest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	var tagged reference.NamedTagged = nil

	opts := make([]distribution.ManifestServiceOption, 0, len(options))
	for _, option := range options {
		opt, ok := option.(distribution.WithTagOption)
		if !ok {
			opts = append(opts, option)
			continue
		}

		ref_tagged, err := reference.WithTag(s.cache.Repository.Named(), opt.Tag)
		if err != nil {
			return nil, fmt.Errorf(`invalid tag "%s": %w`, opt.Tag, err)
		}

		tagged = ref_tagged
	}

	if tagged != nil {
		dgst = digest.Digest("")
		desc, err := s.remote.Tags(ctx).Get(ctx, tagged.Tag())
		if err != nil {
			return nil, fmt.Errorf(`resolve digest for tag"%s" from remote: %w`, tagged.Tag(), err)
		}

		dgst = desc.Digest
	}

	manif, err := s.cache.Get(ctx, dgst)
	if err == nil {
		// Cache hit.
		return manif, nil
	}

	manif, err = s.ManifestService.Get(ctx, dgst, opts...)
	if err != nil {
		return manif, err
	}

	// Update cache.
	s.cache.Set(ctx, dgst, manif)

	return manif, nil
}

type remoteTagService struct {
	distribution.TagService
	cache *TagService
}

func (s *remoteTagService) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	desc, err := s.cache.Get(ctx, tag)
	if err == nil {
		// Cache hit.
		return desc, nil
	}

	desc, err = s.TagService.Get(ctx, tag)
	if err != nil {
		return desc, err
	}

	// Update cache.
	s.cache.Tag(ctx, tag, desc)

	return desc, nil
}

func (s *remoteTagService) All(ctx context.Context) ([]string, error) {
	tags, err := s.cache.All(ctx)
	if err == nil {
		// Cache hit.
		return tags, nil
	}

	tags, err = s.TagService.All(ctx)
	if err != nil {
		return tags, err
	}

	// Update cache.
	s.cache.Set(ctx, tags)

	return tags, nil
}

package registry

import (
	"context"
	"fmt"

	"github.com/distribution/distribution/v3"
	"github.com/opencontainers/go-digest"
)

type ManifestService struct {
	Repo *Repository
}

func (s *ManifestService) Exists(ctx context.Context, dgst digest.Digest) (bool, error) {
	_, ok := s.Repo.Storage.Manifests[dgst.String()]
	return ok, nil
}

func (s *ManifestService) Get(ctx context.Context, dgst digest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	for _, option := range options {
		switch opt := option.(type) {
		case distribution.WithTagOption:
			desc, ok := s.Repo.Storage.Tags[opt.Tag]
			if !ok {
				return nil, ErrNotExists
			}

			dgst = desc.Digest
		}
	}

	manif, ok := s.Repo.Storage.Manifests[dgst.String()]
	if !ok {
		return nil, ErrNotExists
	}

	return manif, nil
}

func (s *ManifestService) Put(ctx context.Context, manifest distribution.Manifest, options ...distribution.ManifestServiceOption) (digest.Digest, error) {
	_, data, err := manifest.Payload()
	if err != nil {
		return "", fmt.Errorf("failed to get payload: %w", err)
	}

	return digest.FromBytes(data), nil
}

func (s *ManifestService) Delete(ctx context.Context, dgst digest.Digest) error {
	delete(s.Repo.Storage.Manifests, dgst.String())
	return nil
}

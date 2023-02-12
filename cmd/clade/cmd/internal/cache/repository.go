package cache

import (
	"context"
	"path/filepath"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
)

type Repository struct {
	Registry  *Registry
	Namespace reference.Named
}

func (r *Repository) Fallback() (*Repository, bool) {
	if r.Registry.Fallback == nil {
		return nil, false
	}

	return &Repository{
		Registry:  r.Registry.Fallback,
		Namespace: r.Namespace,
	}, true
}

func (r *Repository) Path() string {
	return filepath.Join(r.Registry.Root, r.Namespace.Name())
}

func (r *Repository) Named() reference.Named {
	return r.Namespace
}

func (r *Repository) manifests() *ManifestService {
	return &ManifestService{Repository: r}
}

func (r *Repository) Manifests(ctx context.Context, options ...distribution.ManifestServiceOption) (distribution.ManifestService, error) {
	return r.manifests(), nil
}

func (r *Repository) Blobs(ctx context.Context) distribution.BlobStore {
	panic("not implemented")
}

func (r *Repository) tags() *TagService {
	return &TagService{Repository: r}
}

func (r *Repository) Tags(ctx context.Context) distribution.TagService {
	return r.tags()
}

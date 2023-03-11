package registry

import (
	"context"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"golang.org/x/exp/maps"
)

type TagService struct {
	Repo *Repository
}

func (s *TagService) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	desc, ok := s.Repo.Storage.Tags[tag]
	if !ok {
		return distribution.Descriptor{}, errcode.Errors{v2.ErrorCodeManifestUnknown}
	}

	return desc, nil
}

func (s *TagService) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	s.Repo.Storage.Tags[tag] = desc
	return nil
}

func (s *TagService) Untag(ctx context.Context, tag string) error {
	delete(s.Repo.Storage.Tags, tag)
	return nil
}

func (s *TagService) All(ctx context.Context) ([]string, error) {
	return maps.Keys(s.Repo.Storage.Tags), nil
}

func (s *TagService) Lookup(ctx context.Context, digest distribution.Descriptor) ([]string, error) {
	panic("not implemented")
}

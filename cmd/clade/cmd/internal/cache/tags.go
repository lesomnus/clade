package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/rs/zerolog/log"
)

type TagService struct {
	Repository *Repository
}

func (s *TagService) data() map[string]distribution.Descriptor {
	stores := s.Repository.Registry.Tags
	store, ok := stores[s.Repository.Namespace.Name()]
	if !ok {
		store = make(map[string]distribution.Descriptor)
		stores[s.Repository.Namespace.Name()] = store
	}

	return store
}

func (s *TagService) Get(ctx context.Context, tag string) (distribution.Descriptor, error) {
	tagged, err := reference.WithTag(s.Repository.Named(), tag)
	if err != nil {
		return distribution.Descriptor{}, err
	}

	log := log.Ctx(ctx).With().Str("ref", tagged.String()).Str("op", "tags/get").Logger()

	tags := s.data()
	desc, ok := tags[tag]
	if !ok {
		log.Debug().Msg("cache miss")
		return distribution.Descriptor{}, os.ErrNotExist
	}

	log.Debug().Msg("cache hit")
	return desc, nil
}

func (s *TagService) Tag(ctx context.Context, tag string, desc distribution.Descriptor) error {
	tags := s.data()
	tags[tag] = desc
	fmt.Printf("tags: %v\n", tags)

	return nil
}

func (s *TagService) Untag(ctx context.Context, tag string) error {
	tags := s.data()
	delete(tags, tag)

	return nil
}

func (s TagService) PathToAll() string {
	return filepath.Join(s.Repository.Path(), ".tags")
}

func (s TagService) Set(ctx context.Context, tags []string) error {
	dst := s.PathToAll()

	data, err := json.Marshal(tags)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	base := filepath.Dir(dst)
	if err := os.MkdirAll(base, 0755); err != nil {
		return fmt.Errorf(`create directory at "%s": %w`, base, err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf(`write at "%s": %w`, dst, err)
	}

	return nil
}

func (s *TagService) All(ctx context.Context) ([]string, error) {
	tgt := s.PathToAll()
	log := log.Ctx(ctx).With().Str("path", tgt).Str("op", "tags/all").Logger()

	data, err := os.ReadFile(tgt)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Debug().Msg("cache miss")
		} else {
			log.Debug().Err(err).Msg("failed to read cache data")
		}

		return nil, os.ErrNotExist
	}

	tags := make([]string, 0)
	if err := json.Unmarshal(data, &tags); err != nil {
		log.Debug().Err(err).Msg("invalid cache data")

		if err := os.RemoveAll(tgt); err != nil {
			log.Error().Err(err).Msg("failed to remove invalid cache data")
		}
		return nil, os.ErrNotExist
	}

	log.Debug().Msg("cache hit")
	return tags, nil
}

func (s *TagService) Lookup(ctx context.Context, desc distribution.Descriptor) ([]string, error) {
	panic("not implemented")
}

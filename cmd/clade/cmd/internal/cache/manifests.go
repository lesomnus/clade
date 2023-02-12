package cache

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/distribution/distribution/v3"
	"github.com/opencontainers/go-digest"
	"github.com/rs/zerolog/log"
)

type ManifestService struct {
	Repository *Repository
}

func PathToManifest(base string, dgst digest.Digest) string {
	return filepath.Join(base, "manifests", dgst.Algorithm().String(), dgst.Encoded())
}

func (s *ManifestService) PathTo(dgst digest.Digest) string {
	return PathToManifest(s.Repository.Path(), dgst)
}

func (s *ManifestService) Exists(ctx context.Context, dgst digest.Digest) (bool, error) {
	tgt := s.PathTo(dgst)
	log := log.Ctx(ctx).With().Str("path", tgt).Str("op", "manifests/exists").Logger()

	_, err := os.Stat(tgt)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		log.Debug().Msg("failed to stat cache")
		return false, nil
	}

	log.Debug().Msg("cache hit")
	return true, nil
}

func (s *ManifestService) getFromFallback(dgst digest.Digest) ([]byte, error) {
	fallback, ok := s.Repository.Fallback()
	if !ok {
		return nil, os.ErrNotExist
	}

	data, err := os.ReadFile(PathToManifest(fallback.Path(), dgst))
	if err != nil {
		return nil, err
	}

	tgt := s.PathTo(dgst)
	if err := os.MkdirAll(filepath.Dir(tgt), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(tgt, data, 0644); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *ManifestService) Get(ctx context.Context, dgst digest.Digest, options ...distribution.ManifestServiceOption) (distribution.Manifest, error) {
	tgt := s.PathTo(dgst)
	log := log.Ctx(ctx).With().Str("path", tgt).Str("op", "manifests/get").Logger()

	data, err := os.ReadFile(tgt)
	if errors.Is(err, os.ErrNotExist) {
		data, err = s.getFromFallback(dgst)
		if errors.Is(err, os.ErrNotExist) {
			log.Debug().Msg("cache miss")
			return nil, err
		} else if err != nil {
			log.Debug().Msg("failed to restore cache from fallback")
			return nil, err
		}

		log.Debug().Msg("cache restored from fallback")
	} else if err != nil {
		log.Debug().Msg("failed to read cache")
		return nil, err
	}

	defer func() {
		if err == nil {
			return
		}

		log.Debug().Err(err).Msg("invalid cache")
		if err := os.Remove(tgt); err != nil {
			log.Debug().Err(err).Msg("failed to remove invalid cache")
		}
	}()

	sep := bytes.IndexRune(data, '\n')
	if sep < 0 {
		err = errors.New("no media type found")
		return nil, os.ErrNotExist
	}

	manif, _, err := distribution.UnmarshalManifest(string(data[:sep]), data[sep+1:])
	if err != nil {
		return nil, os.ErrNotExist
	}

	log.Debug().Msg("cache hit")
	return manif, nil
}

func (s *ManifestService) Set(ctx context.Context, dgst digest.Digest, manifest distribution.Manifest) error {
	tgt := s.PathTo(dgst)
	log := log.Ctx(ctx).With().Str("path", tgt).Str("op", "manifests/set").Logger()

	media_type, data, err := manifest.Payload()
	if err != nil {
		log.Debug().Err(err).Msg("failed to get payload of manifest")
		return err
	}

	if err := os.MkdirAll(filepath.Dir(tgt), 0755); err != nil {
		log.Debug().Err(err).Msg("failed to create cache directory")
		return err
	}

	f, err := os.OpenFile(tgt, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Debug().Err(err).Msg("failed to get payload of manifest")
		return err
	}
	defer f.Close()

	f.WriteString(media_type)
	f.WriteString("\n")
	f.Write(data)

	log.Debug().Msg("succeeded")
	return nil
}

func (s *ManifestService) Put(ctx context.Context, manifest distribution.Manifest, options ...distribution.ManifestServiceOption) (digest.Digest, error) {
	panic("not implemented")
}

func (s *ManifestService) Delete(ctx context.Context, dgst digest.Digest) error {
	tgt := s.PathTo(dgst)
	log := log.Ctx(ctx).With().Str("path", tgt).Str("op", "manifests/delete").Logger()

	if err := os.Remove(tgt); err != nil {
		log.Debug().Msg("failed to remove cache")
		return nil
	}

	log.Debug().Msg("succeeded")
	return nil
}

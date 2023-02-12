package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
)

type Registry struct {
	Fallback *Registry

	Root string
	Tags map[string]map[string]distribution.Descriptor
}

func NewRegistry(root string) *Registry {
	return &Registry{
		Root: root,
		Tags: make(map[string]map[string]distribution.Descriptor),
	}
}

func ResolveRegistry(base string, at time.Time) (*Registry, error) {
	const format = "2006-01-02"

	src := filepath.Join(base, at.Format(format))
	if err := os.MkdirAll(src, 0755); err != nil {
		return nil, fmt.Errorf(`create directory at "%s": %w`, src, err)
	}

	reg := NewRegistry(src)

	for i := 1; i < 8; i++ {
		past := at.AddDate(0, 0, -i).Format(format)
		history := filepath.Join(base, past)
		if _, err := os.Stat(history); err != nil {
			continue
		}

		// Creates symlink
		// from ./2023-02-11
		// to   /path/to/cache/2023-02-12.fallback
		fallback_name := src + ".fallback"
		if err := os.Symlink(past, fallback_name); err != nil {
			return nil, fmt.Errorf(`symlink from "%s" to "%s": %w`, past, fallback_name, err)
		}

		reg.Fallback = NewRegistry(fallback_name)
		break
	}

	return reg, nil
}

func (r *Registry) repository(named reference.Named) *Repository {
	name_only, err := reference.WithName(named.Name())
	if err != nil {
		panic(err)
	}

	return &Repository{
		Registry:  r,
		Namespace: name_only,
	}
}

func (r *Registry) Repository(named reference.Named) (distribution.Repository, error) {
	return r.repository(named), nil
}

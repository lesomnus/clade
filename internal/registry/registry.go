package registry

import (
	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
)

type Blob struct {
	ContentType string
	Data        []byte
}

func (m *Blob) References() []distribution.Descriptor {
	return []distribution.Descriptor{}
}

func (m *Blob) Payload() (string, []byte, error) {
	return m.ContentType, m.Data, nil
}

type Registry struct {
	Repos map[string]*Repository
}

func NewRegistry() *Registry {
	return &Registry{
		Repos: make(map[string]*Repository),
	}
}

func (r *Registry) Repository(named reference.Named) (distribution.Repository, error) {
	repo, ok := r.Repos[named.Name()]
	if !ok {
		return nil, errcode.Errors{v2.ErrorCodeNameUnknown}
	}

	return repo, nil
}

func (r *Registry) NewRepository(named reference.Named) *Repository {
	repo := NewRepository(named)
	r.Repos[named.Name()] = repo

	return repo
}

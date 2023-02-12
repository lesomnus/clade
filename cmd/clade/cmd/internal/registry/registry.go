package registry

import (
	"os"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
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
	repo, ok := r.Repos[named.String()]
	if !ok {
		return nil, os.ErrNotExist
	}

	return repo, nil
}

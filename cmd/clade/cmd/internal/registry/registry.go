package registry

import (
	"github.com/distribution/distribution/v3"
)

// Implements mock container registry for testing.

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

type RepositoryA struct {
	Name      string
	Manifests map[string]Blob
}

type Registry struct {
	Repos map[string]*Repository
}

func NewRegistry() *Registry {
	return &Registry{
		Repos: make(map[string]*Repository),
	}
}

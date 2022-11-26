package cache

import (
	"github.com/distribution/distribution/reference"
)

type MemCacheStore struct {
	Tags map[string][]string
}

func NewMemCacheStore() *MemCacheStore {
	return &MemCacheStore{
		Tags: make(map[string][]string),
	}
}

func (s *MemCacheStore) Name() string {
	return "@mem"
}

func (s *MemCacheStore) Clear() error {
	s.Tags = make(map[string][]string)
	return nil
}

func (s *MemCacheStore) GetTags(named reference.Named) ([]string, bool) {
	tags, ok := s.Tags[named.Name()]
	return tags, ok
}

func (s *MemCacheStore) SetTags(named reference.Named, tags []string) {
	s.Tags[named.Name()] = tags
}

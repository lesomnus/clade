package cache

import (
	"github.com/distribution/distribution/reference"
)

type NullCacheStore struct{}

func (s *NullCacheStore) Name() string {
	return "@null"
}

func (s *NullCacheStore) Clear() error {
	return nil
}

func (s *NullCacheStore) GetTags(named reference.Named) ([]string, bool) {
	return nil, false
}

func (s *NullCacheStore) SetTags(named reference.Named, tags []string) {

}

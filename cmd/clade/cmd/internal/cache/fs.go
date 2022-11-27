package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/distribution/distribution/reference"
)

type FsCacheStore struct {
	Dir string
}

func (s *FsCacheStore) Name() string {
	return s.Dir
}

func (s *FsCacheStore) Clear() error {
	return os.RemoveAll(filepath.Join(s.Dir, "tags"))
}

func (s *FsCacheStore) GetTags(ref reference.Named) ([]string, bool) {
	tgt := filepath.Join(s.Dir, "tags", ref.Name())
	tags := make([]string, 0)

	data, err := os.ReadFile(tgt)
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal(data, &tags); err != nil {
		os.RemoveAll(tgt)
		return nil, false
	}

	// Log.Trace().Str("path", tgt).Msg("tag cache hit")
	return tags, true
}

func (s *FsCacheStore) SetTags(ref reference.Named, tags []string) {
	tgt := filepath.Join(s.Dir, "tags", ref.Name())
	data, err := json.Marshal(tags)
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(tgt), 0755); err != nil {
		return
	}

	os.WriteFile(tgt, data, 0655)
}

package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/distribution/distribution/reference"
)

var (
	Cache CacheStore
)

type CacheStore struct {
	Dir string
}

func (s *CacheStore) GetTags(ref reference.Named) ([]string, bool) {
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

	return tags, true
}

func (s *CacheStore) SetTags(ref reference.Named, tags []string) {
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

func init() {
	now := time.Now().Format("2006-01-02")

	tmp_dir := filepath.Join(os.TempDir(), "clade-cache-"+now)
	if err := os.MkdirAll(tmp_dir, 0755); err != nil {
		panic(err)
	}

	Cache.Dir = tmp_dir
}

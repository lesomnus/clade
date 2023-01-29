package cache

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/distribution/distribution/reference"
)

type TagCache interface {
	Name() string
	Clear() error
	Get(named reference.Named) ([]string, bool)
	Set(named reference.Named, tags []string)
}

type NullTagCache struct{}

func (c *NullTagCache) Name() string {
	return "@null"
}

func (c *NullTagCache) Clear() error {
	return nil
}

func (c *NullTagCache) Get(named reference.Named) ([]string, bool) {
	return nil, false
}

func (c *NullTagCache) Set(named reference.Named, tags []string) {
}

type MemTagCache struct {
	Tags map[string][]string
}

func NewMemTagCache() *MemTagCache {
	return &MemTagCache{
		Tags: make(map[string][]string),
	}
}

func (c *MemTagCache) Name() string {
	return "@mem"
}

func (c *MemTagCache) Clear() error {
	c.Tags = make(map[string][]string)
	return nil
}

func (c *MemTagCache) Get(named reference.Named) ([]string, bool) {
	tags, ok := c.Tags[named.Name()]
	return tags, ok
}

func (c *MemTagCache) Set(named reference.Named, tags []string) {
	c.Tags[named.Name()] = tags
}

type FsTagCache struct {
	fsCache
}

func NewFsTagCache(dir string) *FsTagCache {
	return &FsTagCache{fsCache: fsCache{Dir: dir}}
}

func (c *FsTagCache) Get(ref reference.Named) ([]string, bool) {
	tgt := filepath.Join(c.Dir, ref.Name())
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

func (c *FsTagCache) Set(ref reference.Named, tags []string) {
	tgt := filepath.Join(c.Dir, ref.Name())
	data, err := json.Marshal(tags)
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(tgt), 0755); err != nil {
		return
	}

	os.WriteFile(tgt, data, 0644)
}

package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Cache is a byte-blob store with per-entry expiry. Implementations must be
// safe for concurrent use.
type Cache interface {
	// Get returns the value for key if present and not expired.
	Get(key string) ([]byte, bool)
	// Set stores val for key, expiring after ttl. A non-positive ttl stores
	// the value without expiry.
	Set(key string, val []byte, ttl time.Duration)
}

// cached decorates a Registry with a Cache. Positive results are cached for the
// configured TTL; ErrNotExist is not cached because a missing target image is
// expected to appear soon (it is about to be built).
type cached struct {
	inner Registry
	cache Cache
	ttl   time.Duration
}

// WithCache wraps a Registry so metadata lookups are served from cache when
// fresh. ttl bounds how long tag listings and image metadata are reused.
func WithCache(inner Registry, cache Cache, ttl time.Duration) Registry {
	return &cached{inner: inner, cache: cache, ttl: ttl}
}

func (c *cached) Tags(ctx context.Context, repo string) ([]string, error) {
	key := "tags:" + repo
	if b, ok := c.cache.Get(key); ok {
		var tags []string
		if err := json.Unmarshal(b, &tags); err == nil {
			return tags, nil
		}
	}

	tags, err := c.inner.Tags(ctx, repo)
	if err != nil {
		return nil, err
	}
	if b, err := json.Marshal(tags); err == nil {
		c.cache.Set(key, b, c.ttl)
	}
	return tags, nil
}

func (c *cached) Stat(ctx context.Context, ref string) (*ImageInfo, error) {
	key := "stat:" + ref
	if b, ok := c.cache.Get(key); ok {
		var info ImageInfo
		if err := json.Unmarshal(b, &info); err == nil {
			return &info, nil
		}
	}

	info, err := c.inner.Stat(ctx, ref)
	if err != nil {
		return nil, err
	}
	if b, err := json.Marshal(info); err == nil {
		c.cache.Set(key, b, c.ttl)
	}
	return info, nil
}

// memEntry is a cached value with its expiry.
type memEntry struct {
	val       []byte
	expiresAt time.Time // zero means no expiry
}

// MemCache is an in-memory Cache.
type MemCache struct {
	mu      sync.Mutex
	entries map[string]memEntry
	now     func() time.Time
}

// NewMemCache creates an empty in-memory cache.
func NewMemCache() *MemCache {
	return &MemCache{entries: map[string]memEntry{}, now: time.Now}
}

func (c *MemCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if !e.expiresAt.IsZero() && c.now().After(e.expiresAt) {
		delete(c.entries, key)
		return nil, false
	}
	return e.val, true
}

func (c *MemCache) Set(key string, val []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = c.now().Add(ttl)
	}
	c.entries[key] = memEntry{val: val, expiresAt: exp}
}

// fileEntry is the on-disk representation of a cached value.
type fileEntry struct {
	ExpiresAt time.Time `json:"expires_at"`
	Val       []byte    `json:"val"`
}

// FileCache is a filesystem-backed Cache. It persists entries across runs,
// which lets CI reuse metadata and conserve registry rate limits.
type FileCache struct {
	dir string
	now func() time.Time
}

// NewFileCache creates a cache rooted at dir, creating it if needed.
func NewFileCache(dir string) (*FileCache, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir %q: %w", dir, err)
	}
	return &FileCache{dir: dir, now: time.Now}, nil
}

func (c *FileCache) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, hex.EncodeToString(sum[:])+".json")
}

func (c *FileCache) Get(key string) ([]byte, bool) {
	b, err := os.ReadFile(c.path(key))
	if err != nil {
		return nil, false
	}

	var e fileEntry
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, false
	}
	if !e.ExpiresAt.IsZero() && c.now().After(e.ExpiresAt) {
		_ = os.Remove(c.path(key))
		return nil, false
	}
	return e.Val, true
}

func (c *FileCache) Set(key string, val []byte, ttl time.Duration) {
	e := fileEntry{Val: val}
	if ttl > 0 {
		e.ExpiresAt = c.now().Add(ttl)
	}
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	_ = os.WriteFile(c.path(key), b, 0o644)
}

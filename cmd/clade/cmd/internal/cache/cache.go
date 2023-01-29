package cache

import (
	"os"
	"path/filepath"
	"time"
)

var Cache CacheStore

type CacheStore struct {
	Manifests ManifestCache
	Tags      TagCache
}

func init() {
	now := time.Now().Format("2006-01-02")

	dir, ok := os.LookupEnv("CLADE_CACHE_DIR")
	if !ok {
		dir = filepath.Join(os.TempDir(), "clade-cache-"+now)
	}

	Cache.Manifests = NewFsManifestCache(filepath.Join(dir, "manifests")) // TODO: cache monthly?
	Cache.Tags = NewFsTagCache(filepath.Join(dir, "tags"))
}

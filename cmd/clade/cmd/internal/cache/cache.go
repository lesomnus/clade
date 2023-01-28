package cache

import (
	"os"
	"path/filepath"
	"time"
)

var Cache CacheStore

type CacheStore struct {
	Tags      TagCache
	Manifests ManifestCache
}

func init() {
	now := time.Now().Format("2006-01-02")

	dir, ok := os.LookupEnv("CLADE_CACHE_DIR")
	if !ok {
		dir = filepath.Join(os.TempDir(), "clade-cache-"+now)
	}

	Cache.Manifests = NewMemManifestCache()
	Cache.Tags = &FsTagCache{Dir: filepath.Join(dir, "tags")}
}

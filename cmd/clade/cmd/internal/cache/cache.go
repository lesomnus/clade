package cache

import (
	"os"
	"path/filepath"
	"time"

	"github.com/distribution/distribution/reference"
)

var Cache CacheStore

type CacheStore interface {
	Name() string
	Clear() error
	GetTags(named reference.Named) ([]string, bool)
	SetTags(named reference.Named, tags []string)
}

func init() {
	now := time.Now().Format("2006-01-02")

	dir, ok := os.LookupEnv("CLADE_CACHE_DIR")
	if !ok {
		dir = filepath.Join(os.TempDir(), "clade-cache-"+now)
	}

	Cache = &FsCacheStore{Dir: dir}
}

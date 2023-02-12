package cmd

import (
	"os"
	"path/filepath"
	"time"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/spf13/cobra"
)

var RegistryCache *cache.Registry

func resolveRegistryCache() *cache.Registry {
	if RegistryCache != nil {
		return RegistryCache
	}

	now := time.Now()

	dir, ok := os.LookupEnv("CLADE_CACHE_DIR")
	if !ok {
		dir = filepath.Join(os.TempDir(), "clade-cache")
	}

	reg, err := cache.ResolveRegistry(dir, now)
	if err != nil {
		panic(err)
	} else {
		RegistryCache = reg
	}

	return RegistryCache
}

var cache_cmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage caches",
}

func init() {
	resolveRegistryCache()

	root_cmd.AddCommand(cache_cmd)
}

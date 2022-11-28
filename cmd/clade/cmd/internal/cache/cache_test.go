package cache_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestCacheName(t *testing.T) {
	t.Run("name of fs cache is absolute path of directory", func(t *testing.T) {
		require := require.New(t)
		require.True(filepath.IsAbs(cache.Cache.Name()))
	})

	t.Run("name not fs cache store starts with @", func(t *testing.T) {
		tcs := []struct {
			desc  string
			store cache.CacheStore
		}{
			{
				desc:  "mem",
				store: cache.NewMemCacheStore(),
			},
			{
				desc:  "null",
				store: &cache.NullCacheStore{},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)
				require.True(strings.HasPrefix(tc.store.Name(), "@"))
			})
		}
	})
}

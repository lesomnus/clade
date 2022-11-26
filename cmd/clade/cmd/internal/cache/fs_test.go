package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestFsCacheStoreTags(t *testing.T) {
	t.Run("overwrite", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()
		cache := cache.FsCacheStore{Dir: tmp}

		named, err := reference.ParseNamed("ghcr.io/repo/name")
		require.NoError(err)

		{
			tags := []string{"foo", "bar", "baz"}
			cache.SetTags(named, tags)
			tags_read, ok := cache.GetTags(named)
			require.True(ok)
			require.Equal(tags, tags_read)
		}

		{
			tags := []string{"a", "b", "c"}
			cache.SetTags(named, tags)
			tags_read, ok := cache.GetTags(named)
			require.True(ok)
			require.Equal(tags, tags_read)
		}
	})

	t.Run("not fails if there is no directory", func(t *testing.T) {
		require := require.New(t)

		cache := cache.FsCacheStore{Dir: "/not exists"}

		named, err := reference.ParseNamed("ghcr.io/repo/name")
		require.NoError(err)

		cache.SetTags(named, []string{"foo", "bar", "baz"})
		_, ok := cache.GetTags(named)
		require.False(ok)
	})

	t.Run("not fails if data is invalid", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()
		cache := cache.FsCacheStore{Dir: tmp}

		named, err := reference.ParseNamed("ghcr.io/repo/name")
		require.NoError(err)

		cache.SetTags(named, []string{"foo", "bar", "baz"})
		os.WriteFile(filepath.Join(tmp, "tags", "ghcr.io/repo/name"), []byte("foo"), 0644)

		_, ok := cache.GetTags(named)
		require.False(ok)
	})

	t.Run("clear only removes its content not the directory", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()
		cache := cache.FsCacheStore{Dir: tmp}

		named, err := reference.ParseNamed("ghcr.io/repo/name")
		require.NoError(err)

		cache.SetTags(named, []string{"foo", "bar", "baz"})
		cache.Clear()

		_, err = os.Stat(tmp)
		require.NoError(err)
	})

}

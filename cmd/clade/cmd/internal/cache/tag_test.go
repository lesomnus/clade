package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func testTagCache(t *testing.T, get func() cache.TagCache) {
	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	t.Run("get fail if not exists", func(t *testing.T) {
		require := require.New(t)
		c := get()

		_, ok := c.Get(named)
		require.False(ok)
	})

	t.Run("retrieve", func(t *testing.T) {
		require := require.New(t)
		c := get()

		c.Set(named, []string{"foo", "bar", "baz"})
		tags, ok := c.Get(named)
		require.True(ok)
		require.ElementsMatch([]string{"foo", "bar", "baz"}, tags)
	})

	t.Run("overwrite", func(t *testing.T) {
		require := require.New(t)
		c := get()

		c.Set(named, []string{"foo", "bar", "baz"})
		c.Set(named, []string{"a", "b", "c"})
		tags, ok := c.Get(named)
		require.True(ok)
		require.ElementsMatch([]string{"a", "b", "c"}, tags)
	})

	t.Run("clear", func(t *testing.T) {
		require := require.New(t)
		c := get()

		c.Set(named, []string{"a", "b", "c"})
		c.Clear()
		_, ok := c.Get(named)
		require.False(ok)
	})
}

func TestMemTagCache(t *testing.T) {
	get := func() cache.TagCache {
		return cache.NewMemTagCache()
	}

	c := get()
	require.Equal(t, "@mem", c.Name())

	testTagCache(t, get)
}

func TestFsTagCache(t *testing.T) {
	get := func() cache.TagCache {
		tmp := t.TempDir()
		return &cache.FsTagCache{Dir: tmp}
	}

	testTagCache(t, get)

	t.Run("name is the path of where the cache stored", func(t *testing.T) {
		c := cache.FsTagCache{Dir: "/path/to/cache"}
		require.Equal(t, c.Dir, c.Name())
	})

	t.Run("not fails if there is no directory", func(t *testing.T) {
		require := require.New(t)

		c := cache.FsTagCache{Dir: "/not exists"}

		named, err := reference.ParseNamed("cr.io/repo/name")
		require.NoError(err)

		c.Set(named, []string{"foo", "bar", "baz"})
		_, ok := c.Get(named)
		require.False(ok)
	})

	t.Run("not fails even if data is invalid", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()
		c := cache.FsTagCache{Dir: tmp}

		named, err := reference.ParseNamed("cr.io/repo/name")
		require.NoError(err)

		c.Set(named, []string{"foo", "bar", "baz"})
		os.WriteFile(filepath.Join(tmp, "cr.io/repo/name"), []byte("foo"), 0644)

		_, ok := c.Get(named)
		require.False(ok)
	})

	t.Run("clear only removes its content not the directory", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()
		c := cache.FsTagCache{Dir: tmp}

		named, err := reference.ParseNamed("cr.io/repo/name")
		require.NoError(err)

		c.Set(named, []string{"foo", "bar", "baz"})
		c.Clear()

		_, err = os.Stat(tmp)
		require.NoError(err)
	})
}

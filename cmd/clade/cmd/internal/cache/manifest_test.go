package cache_test

import (
	"os"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func testManifestCache(t *testing.T, get func() cache.ManifestCache) {
	digest := digest.NewDigestFromEncoded(digest.Canonical, "something")
	manif_foo := &registry.Blob{ContentType: "testing", Data: []byte("foo")}
	manif_bar := &registry.Blob{ContentType: "testing", Data: []byte("bar")}

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)
	tagged, err := reference.WithTag(named, "tag")
	require.NoError(t, err)

	t.Run("by reference", func(t *testing.T) {
		t.Run("get fail if not exists", func(t *testing.T) {
			require := require.New(t)
			c := get()

			_, ok := c.GetByRef(tagged)
			require.False(ok)
		})

		t.Run("retrieve", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(tagged, manif_foo)
			manif, ok := c.GetByRef(tagged)
			require.True(ok)

			_, data, _ := manif.Payload()
			require.Equal([]byte("foo"), data)
		})

		t.Run("overwrite", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(tagged, manif_foo)
			c.SetByRef(tagged, manif_bar)
			manif, ok := c.GetByRef(tagged)
			require.True(ok)

			_, data, _ := manif.Payload()
			require.Equal([]byte("bar"), data)
		})

		t.Run("clear", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(tagged, manif_foo)
			c.Clear()
			_, ok := c.GetByRef(tagged)
			require.False(ok)
		})
	})

	t.Run("by digest", func(t *testing.T) {
		t.Run("get fail if not exists", func(t *testing.T) {
			require := require.New(t)
			c := get()

			_, ok := c.GetByDigest("plain:not_exists")
			require.False(ok)
		})

		t.Run("retrieve", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, manif_foo)
			manif, ok := c.GetByDigest(digest)
			require.True(ok)

			_, data, _ := manif.Payload()
			require.Equal([]byte("foo"), data)
		})

		t.Run("overwrite", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, manif_foo)
			c.SetByDigest(digest, manif_bar)
			manif, ok := c.GetByDigest(digest)
			require.True(ok)

			_, data, _ := manif.Payload()
			require.Equal([]byte("bar"), data)
		})

		t.Run("clear", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, manif_foo)
			c.Clear()
			_, ok := c.GetByDigest(digest)
			require.False(ok)
		})
	})
}

func TestMemManifestCache(t *testing.T) {
	require.Equal(t, "@mem", cache.NewMemManifestCache().Name())

	testManifestCache(t, func() cache.ManifestCache {
		return cache.NewMemManifestCache()
	})
}

func TestFsManifestCache(t *testing.T) {
	get := func() *cache.FsManifestCache {
		tmp := t.TempDir()
		return cache.NewFsManifestCache(tmp)
	}

	testManifestCache(t, func() cache.ManifestCache { return get() })

	dgst := digest.NewDigestFromEncoded(digest.Canonical, "something")
	manif := &registry.Blob{ContentType: "testing", Data: []byte("foo")}

	t.Run("name is the path of where the cache stored", func(t *testing.T) {
		c := cache.NewFsTagCache("/path/to/cache")
		require.Equal(t, c.Dir, c.Name())
	})

	t.Run("not fails even if there is no directory", func(t *testing.T) {
		require := require.New(t)

		c := cache.NewFsManifestCache("")
		c.Dir = "/not exists"
		c.SetByDigest(dgst, manif)
		_, ok := c.GetByDigest(dgst)
		require.False(ok)
	})

	t.Run("not fails even if data is invalid", func(t *testing.T) {
		require := require.New(t)

		c := get()
		c.SetByDigest(dgst, manif)
		os.WriteFile(c.ToPath(dgst), []byte("not registered type\nsome data"), 0644)

		_, ok := c.GetByDigest(dgst)
		require.False(ok)
	})

	t.Run("clear only removes its content not the directory", func(t *testing.T) {
		require := require.New(t)

		c := get()
		c.SetByDigest(dgst, manif)
		c.Clear()

		_, err := os.Stat(c.Name())
		require.NoError(err)
	})
}

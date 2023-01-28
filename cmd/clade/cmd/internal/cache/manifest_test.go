package cache_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func testManifestCache(t *testing.T, get func() cache.ManifestCache) {
	digest := digest.NewDigestFromEncoded(digest.Canonical, "something")
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

			c.SetByRef(tagged, &registry.Manifest{ContentType: "foo"})
			manif, ok := c.GetByRef(tagged)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("foo", media_type)
		})

		t.Run("overwrite", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(tagged, &registry.Manifest{ContentType: "foo"})
			c.SetByRef(tagged, &registry.Manifest{ContentType: "bar"})
			manif, ok := c.GetByRef(tagged)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("bar", media_type)
		})

		t.Run("clear", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(tagged, &registry.Manifest{ContentType: "foo"})
			c.Clear()
			_, ok := c.GetByRef(tagged)
			require.False(ok)
		})
	})

	t.Run("by digest", func(t *testing.T) {
		t.Run("get fail if not exists", func(t *testing.T) {
			require := require.New(t)
			c := get()

			_, ok := c.GetByDigest("not exists")
			require.False(ok)
		})

		t.Run("retrieve", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, &registry.Manifest{ContentType: "foo"})
			manif, ok := c.GetByDigest(digest)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("foo", media_type)
		})

		t.Run("overwrite", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, &registry.Manifest{ContentType: "foo"})
			c.SetByDigest(digest, &registry.Manifest{ContentType: "bar"})
			manif, ok := c.GetByDigest(digest)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("bar", media_type)
		})

		t.Run("clear", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, &registry.Manifest{ContentType: "foo"})
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

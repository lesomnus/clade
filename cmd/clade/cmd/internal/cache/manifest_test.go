package cache_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/distribution/distribution/v3"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

type typedManifest struct {
	media_type string
}

func (m *typedManifest) References() []distribution.Descriptor {
	return []distribution.Descriptor{}
}

func (m *typedManifest) Payload() (string, []byte, error) {
	return m.media_type, []byte{}, nil
}

func testManifestCache(t *testing.T, get func() cache.ManifestCache) {
	digest := digest.NewDigestFromEncoded(digest.Canonical, "something")
	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	t.Run("by reference", func(t *testing.T) {
		t.Run("get fail if not exists", func(t *testing.T) {
			require := require.New(t)
			c := get()

			_, ok := c.GetByRef(named)
			require.False(ok)
		})

		t.Run("retrieve", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(named, &typedManifest{media_type: "foo"})
			manif, ok := c.GetByRef(named)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("foo", media_type)
		})

		t.Run("overwrite", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(named, &typedManifest{media_type: "foo"})
			c.SetByRef(named, &typedManifest{media_type: "bar"})
			manif, ok := c.GetByRef(named)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("bar", media_type)
		})

		t.Run("clear", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByRef(named, &typedManifest{media_type: "foo"})
			c.Clear()
			_, ok := c.GetByRef(named)
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

			c.SetByDigest(digest, &typedManifest{media_type: "foo"})
			manif, ok := c.GetByDigest(digest)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("foo", media_type)
		})

		t.Run("overwrite", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, &typedManifest{media_type: "foo"})
			c.SetByDigest(digest, &typedManifest{media_type: "bar"})
			manif, ok := c.GetByDigest(digest)
			require.True(ok)

			media_type, _, _ := manif.Payload()
			require.Equal("bar", media_type)
		})

		t.Run("clear", func(t *testing.T) {
			require := require.New(t)
			c := get()

			c.SetByDigest(digest, &typedManifest{media_type: "foo"})
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

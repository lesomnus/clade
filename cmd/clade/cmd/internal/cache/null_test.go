package cache_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestNullTagCache(t *testing.T) {
	require := require.New(t)

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(err)

	c := cache.NullTagCache{}
	require.Equal("@null", c.Name())

	c.Set(named, []string{})
	_, ok := c.Get(named)
	require.False(ok)
}

func TestNullManifestCache(t *testing.T) {
	require := require.New(t)

	digest := digest.NewDigestFromEncoded(digest.Canonical, "something")
	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(err)

	c := cache.NullManifestCache{}
	require.Equal("@null", c.Name())

	c.SetByDigest(digest, &typedManifest{})
	_, ok := c.GetByDigest(digest)
	require.False(ok)

	c.SetByRef(named, &typedManifest{})
	_, ok = c.GetByRef(named)
	require.False(ok)
}

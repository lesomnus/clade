package cache_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestNullCacheStore(t *testing.T) {
	c := cache.NullCacheStore{}

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	t.Run("Clear", func(t *testing.T) {
		require := require.New(t)

		err := c.Clear()
		require.NoError(err)
	})

	t.Run("GetTags", func(t *testing.T) {
		require := require.New(t)

		_, ok := c.GetTags(named)
		require.False(ok)
	})

	t.Run("SetTags", func(t *testing.T) {
		c.SetTags(named, nil)
	})
}

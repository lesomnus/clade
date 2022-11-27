package cache_test

import (
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestMemCacheStoreTags(t *testing.T) {
	require := require.New(t)

	named, err := reference.ParseNamed("ghcr.io/repo/name")
	require.NoError(err)

	cache := cache.NewMemCacheStore()

	_, ok := cache.GetTags(named)
	require.False(ok)

	cache.SetTags(named, []string{"foo", "bar", "baz"})
	tags, ok := cache.GetTags(named)
	require.True(ok)
	require.ElementsMatch([]string{"foo", "bar", "baz"}, tags)

	cache.SetTags(named, []string{"a", "b", "c"})
	tags, ok = cache.GetTags(named)
	require.True(ok)
	require.ElementsMatch([]string{"a", "b", "c"}, tags)

	cache.Clear()
	_, ok = cache.GetTags(named)
	require.False(ok)
}

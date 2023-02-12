package cache_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestTagsAll(t *testing.T) {
	ctx := context.Background()

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	withSvc := func(tester func(t *testing.T, tags *cache.TagService)) func(*testing.T) {
		return func(t *testing.T) {
			tmp := t.TempDir()
			reg := cache.NewRegistry(tmp)

			repo, err := reg.Repository(named)
			require.NoError(t, err)

			tags, ok := repo.Tags(ctx).(*cache.TagService)
			require.True(t, ok)

			tester(t, tags)
		}
	}

	t.Run("Set stores tag lits", withSvc(func(t *testing.T, tags *cache.TagService) {
		require := require.New(t)

		_, err = tags.All(ctx)
		require.Error(err)

		err = tags.Set(ctx, []string{"foo", "bar"})
		require.NoError(err)

		tags_all, err := tags.All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"foo", "bar"}, tags_all)
	}))

	t.Run("cache is removed if it is invalid", withSvc(func(t *testing.T, tags *cache.TagService) {
		require := require.New(t)

		err := os.MkdirAll(filepath.Dir(tags.PathToAll()), 0755)
		require.NoError(err)

		err = os.WriteFile(tags.PathToAll(), []byte("invalid data"), 0644)
		require.NoError(err)

		_, err = tags.All(ctx)
		require.ErrorIs(err, os.ErrNotExist)

		_, err = os.Stat(tags.PathToAll())
		require.ErrorIs(err, os.ErrNotExist)
	}))
}

func TestTags(t *testing.T) {
	require := require.New(t)

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(err)

	tmp := t.TempDir()
	reg := cache.NewRegistry(tmp)

	repo, err := reg.Repository(named)
	require.NoError(err)

	ctx := context.Background()
	tags := repo.Tags(ctx)

	_, err = tags.Get(ctx, "foo")
	require.ErrorIs(err, os.ErrNotExist)

	desc := distribution.Descriptor{Size: 42}
	err = tags.Tag(ctx, "foo", desc)
	require.NoError(err)

	desc_loaded, err := tags.Get(ctx, "foo")
	require.NoError(err)
	require.Equal(desc, desc_loaded)

	err = tags.Untag(ctx, "foo")
	require.NoError(err)

	_, err = tags.Get(ctx, "foo")
	require.ErrorIs(err, os.ErrNotExist)

	err = tags.Untag(ctx, "foo")
	require.NoError(err)
}

package registry_test

import (
	"context"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestManifestService(t *testing.T) {
	ctx := context.Background()

	named, err := reference.WithName("cr.io/repo/name")
	require.NoError(t, err)

	test := func(tester func(*testing.T, *registry.Repository, distribution.ManifestService)) func(t *testing.T) {
		return func(t *testing.T) {
			repo := registry.NewRepository(named)
			ms, err := repo.Manifests(ctx)
			require.NoError(t, err)

			tester(t, repo, ms)
		}
	}

	t.Run("Exists", test(func(t *testing.T, repo *registry.Repository, ms distribution.ManifestService) {
		require := require.New(t)

		desc, _ := repo.PopulateManifest()
		ok, err := ms.Exists(ctx, desc.Digest)
		require.NoError(err)
		require.True(ok)
	}))

	t.Run("Get", test(func(t *testing.T, repo *registry.Repository, ms distribution.ManifestService) {
		require := require.New(t)

		desc, manif := repo.PopulateManifest()
		manif_actual, err := ms.Get(ctx, desc.Digest)
		require.NoError(err)
		require.Equal(manif, manif_actual)
	}))

	t.Run("Get by tag", test(func(t *testing.T, repo *registry.Repository, ms distribution.ManifestService) {
		require := require.New(t)

		tagged, _, manif := repo.PopulateImage()
		manif_actual, err := ms.Get(ctx, digest.Digest(""), distribution.WithTag(tagged.Tag()))
		require.NoError(err)
		require.Equal(manif, manif_actual)
	}))

	t.Run("Put", test(func(t *testing.T, repo *registry.Repository, ms distribution.ManifestService) {
		require := require.New(t)

		desc, manif := repo.PopulateManifest()
		err := ms.Delete(ctx, desc.Digest)
		require.NoError(err)

		dgst, err := ms.Put(ctx, manif)
		require.NoError(err)
		require.Equal(desc.Digest, dgst)
	}))

	t.Run("Delete", test(func(t *testing.T, repo *registry.Repository, ms distribution.ManifestService) {
		require := require.New(t)

		desc, _ := repo.PopulateManifest()
		err := ms.Delete(ctx, desc.Digest)
		require.NoError(err)

		ok, err := ms.Exists(ctx, desc.Digest)
		require.NoError(err)
		require.False(ok)
	}))
}

func TestTagService(t *testing.T) {
	ctx := context.Background()

	named, err := reference.WithName("cr.io/repo/name")
	require.NoError(t, err)

	test := func(tester func(*testing.T, *registry.Repository, distribution.TagService)) func(t *testing.T) {
		return func(t *testing.T) {
			repo := registry.NewRepository(named)
			ts := repo.Tags(ctx)

			tester(t, repo, ts)
		}
	}

	t.Run("Get", test(func(t *testing.T, repo *registry.Repository, ts distribution.TagService) {
		require := require.New(t)

		tagged, desc, _ := repo.PopulateImage()
		desc_actual, err := ts.Get(ctx, tagged.Tag())
		require.NoError(err)
		require.Equal(desc, desc_actual)
	}))

	t.Run("Tag", test(func(t *testing.T, repo *registry.Repository, ts distribution.TagService) {
		require := require.New(t)

		_, err := ts.Get(ctx, "foo")
		require.ErrorIs(err, registry.ErrNotExists)

		desc, _ := repo.PopulateManifest()
		err = ts.Tag(ctx, "foo", desc)
		require.NoError(err)

		desc_actual, err := ts.Get(ctx, "foo")
		require.NoError(err)
		require.Equal(desc, desc_actual)
	}))

	t.Run("Untag", test(func(t *testing.T, repo *registry.Repository, ts distribution.TagService) {
		require := require.New(t)

		tagged, _, _ := repo.PopulateImage()
		err := ts.Untag(ctx, tagged.Tag())
		require.NoError(err)

		_, err = ts.Get(ctx, tagged.Tag())
		require.ErrorIs(err, registry.ErrNotExists)
	}))

	t.Run("All", test(func(t *testing.T, repo *registry.Repository, ts distribution.TagService) {
		require := require.New(t)

		tagged_a, _, _ := repo.PopulateImage()
		tagged_b, _, _ := repo.PopulateImage()
		tags_all, err := ts.All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{tagged_a.Tag(), tagged_b.Tag()}, tags_all)
	}))
}

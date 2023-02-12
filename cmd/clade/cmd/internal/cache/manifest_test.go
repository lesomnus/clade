package cache_test

import (
	"context"
	"os"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestManifests(t *testing.T) {
	ctx := context.Background()

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	withSvc := func(tester func(t *testing.T, manifests *cache.ManifestService)) func(*testing.T) {
		return func(t *testing.T) {
			require := require.New(t)

			tmp := t.TempDir()
			reg := cache.NewRegistry(tmp)

			repo, err := reg.Repository(named)
			require.NoError(err)

			svc, err := repo.Manifests(ctx)
			require.NoError(err)

			manifests, ok := svc.(*cache.ManifestService)
			require.True(ok)

			tester(t, manifests)
		}
	}

	dgst := digest.FromString("something")
	manif, err := schema2.FromStruct(schema2.Manifest{
		Versioned: schema2.SchemaVersion,
		Layers:    []distribution.Descriptor{{Size: 42}},
	})
	require.NoError(t, err)

	t.Run("Set", withSvc(func(t *testing.T, manifests *cache.ManifestService) {
		require := require.New(t)

		ok, err := manifests.Exists(ctx, dgst)
		require.NoError(err)
		require.False(ok)

		_, err = manifests.Get(ctx, dgst)
		require.ErrorIs(err, os.ErrNotExist)

		err = manifests.Set(ctx, dgst, manif)
		require.NoError(err)

		ok, err = manifests.Exists(ctx, dgst)
		require.NoError(err)
		require.True(ok)

		manif_loaded, err := manifests.Get(ctx, dgst)
		require.NoError(err)
		require.Equal(manif, manif_loaded)
	}))

	t.Run("Delete", withSvc(func(t *testing.T, manifests *cache.ManifestService) {
		require := require.New(t)

		err := manifests.Set(ctx, dgst, manif)
		require.NoError(err)

		err = manifests.Delete(ctx, dgst)
		require.NoError(err)

		ok, err := manifests.Exists(ctx, dgst)
		require.NoError(err)
		require.False(ok)

		_, err = manifests.Get(ctx, dgst)
		require.ErrorIs(err, os.ErrNotExist)
	}))

	t.Run("Get from fallback", withSvc(func(t *testing.T, fallback *cache.ManifestService) {
		require := require.New(t)

		err := fallback.Set(ctx, dgst, manif)
		require.NoError(err)

		withSvc(func(t *testing.T, manifests *cache.ManifestService) {
			manifests.Repository.Registry.Fallback = fallback.Repository.Registry

			manif_loaded, err := manifests.Get(ctx, dgst)
			require.NoError(err)
			require.Equal(manif, manif_loaded)

			manifests.Repository.Registry.Fallback = nil
			manif_loaded, err = manifests.Get(ctx, dgst)
			require.NoError(err)
			require.Equal(manif, manif_loaded)
		})(t)
	}))

	t.Run("invalid cache data is removed", withSvc(func(t *testing.T, manifests *cache.ManifestService) {
		require := require.New(t)

		err := manifests.Set(ctx, dgst, manif)
		require.NoError(err)

		manif_path := manifests.PathTo(dgst)

		err = os.WriteFile(manif_path, []byte("invalid data"), 0644)
		require.NoError(err)

		_, err = manifests.Get(ctx, dgst)
		require.Error(err)

		_, err = os.Stat(manif_path)
		require.Error(err, os.ErrNotExist)
	}))
}

package cache_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestRegistryResolve(t *testing.T) {
	t.Run("finds fallback", func(t *testing.T) {
		require := require.New(t)

		now := time.Now()
		yesterday := now.AddDate(0, 0, -1)

		tmp := t.TempDir()
		fallback, err := cache.ResolveRegistry(tmp, yesterday)
		require.NoError(err)

		reg, err := cache.ResolveRegistry(tmp, now)
		require.NoError(err)
		require.NotNil(reg.Fallback)

		t.Run("should not fail when resolve again", func(t *testing.T) {
			reg_new, err := cache.ResolveRegistry(tmp, now)
			require.NoError(err)
			require.NotNil(reg.Fallback)
			require.Equal(reg.Fallback.Root, reg_new.Fallback.Root)
		})

		// Get cache from fallback.
		named, err := reference.ParseNamed("cr.io/repo/name")
		require.NoError(err)

		dgst := digest.FromString("something")
		manif, err := schema2.FromStruct(schema2.Manifest{
			Versioned: schema2.SchemaVersion,
			Layers:    []distribution.Descriptor{{Size: 42}},
		})
		require.NoError(err)

		fallback_repo, err := fallback.Repository(named)
		require.NoError(err)

		ctx := context.Background()
		{
			manifests, err := fallback_repo.Manifests(ctx)
			require.NoError(err)

			fallback_manifests, ok := manifests.(*cache.ManifestService)
			require.True(ok)

			err = fallback_manifests.Set(ctx, dgst, manif)
			require.NoError(err)
		}

		repo, err := reg.Repository(named)
		require.NoError(err)

		manifests, err := repo.Manifests(ctx)
		require.NoError(err)

		manif_loaded, err := manifests.Get(ctx, dgst)
		require.NoError(err)
		require.Equal(manif, manif_loaded)
	})

	t.Run("fails if fallback is not a directory", func(t *testing.T) {
		require := require.New(t)

		now := time.Now()
		tmp := t.TempDir()

		fallback_name := filepath.Join(tmp, now.Format("2006-01-02")+".fallback")
		err := os.WriteFile(fallback_name, []byte("42"), 0644)
		require.NoError(err)

		_, err = cache.ResolveRegistry(tmp, now)
		require.ErrorContains(err, "must be a directory")
	})
}

package cache_test

import (
	"context"
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
	require := require.New(t)

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)

	tmp := t.TempDir()
	fallback, err := cache.ResolveRegistry(tmp, yesterday)
	require.NoError(err)

	reg, err := cache.ResolveRegistry(tmp, now)
	require.NoError(err)
	require.NotNil(reg.Fallback)

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
}

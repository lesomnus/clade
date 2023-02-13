package cache_test

import (
	"context"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestRemoteManifests(t *testing.T) {
	ctx := context.Background()

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	withSvc := func(tester func(t *testing.T, local *cache.Repository, remote *registry.Repository, manifests distribution.ManifestService)) func(*testing.T) {
		return func(t *testing.T) {
			require := require.New(t)

			tmp := t.TempDir()
			local := cache.NewRegistry(tmp)
			local_repo, err := local.Repository(named)
			require.NoError(err)

			remote := registry.NewRegistry()
			remote_repo := remote.NewRepository(named)

			reg := cache.WithRemote(local, remote)
			repo, err := reg.Repository(named)
			require.NoError(err)

			manifests, err := repo.Manifests(ctx)
			require.NoError(err)

			tester(t, local_repo.(*cache.Repository), remote_repo, manifests)
		}
	}

	t.Run("Get by digest", withSvc(func(t *testing.T, local *cache.Repository, remote *registry.Repository, manifests distribution.ManifestService) {
		require := require.New(t)

		desc, manif := remote.PopulateManifest()
		manif_fetched, err := manifests.Get(ctx, desc.Digest)
		require.NoError(err)
		require.Equal(manif, manif_fetched)

		err = manifests.Delete(ctx, desc.Digest)
		require.NoError(err)

		ok, err := manifests.Exists(ctx, desc.Digest)
		require.NoError(err)
		require.True(ok)

		manif_loaded, err := manifests.Get(ctx, desc.Digest)
		require.NoError(err)
		require.Equal(manif, manif_loaded)
	}))

	t.Run("Get by tag", withSvc(func(t *testing.T, local *cache.Repository, remote *registry.Repository, manifests distribution.ManifestService) {
		require := require.New(t)

		tagged, _, manif := remote.PopulateImage()
		manif_fetched, err := manifests.Get(ctx, digest.Digest(""), distribution.WithTag(tagged.Tag()))
		require.NoError(err)
		require.Equal(manif, manif_fetched)

		err = remote.Tags(ctx).Untag(ctx, tagged.Tag())
		require.NoError(err)

		manif_loaded, err := manifests.Get(ctx, digest.Digest(""), distribution.WithTag(tagged.Tag()))
		require.NoError(err)
		require.Equal(manif, manif_loaded)
	}))
}

func TestRemoteTags(t *testing.T) {
	ctx := context.Background()

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	withSvc := func(tester func(t *testing.T, local *cache.Repository, remote *registry.Repository, tags distribution.TagService)) func(*testing.T) {
		return func(t *testing.T) {
			require := require.New(t)

			tmp := t.TempDir()
			local := cache.NewRegistry(tmp)
			local_repo, err := local.Repository(named)
			require.NoError(err)

			remote := registry.NewRegistry()
			remote_repo := remote.NewRepository(named)

			reg := cache.WithRemote(local, remote)
			repo, err := reg.Repository(named)
			require.NoError(err)

			tags := repo.Tags(ctx)

			tester(t, local_repo.(*cache.Repository), remote_repo, tags)
		}
	}

	t.Run("Get", withSvc(func(t *testing.T, local *cache.Repository, remote *registry.Repository, tags distribution.TagService) {
		require := require.New(t)

		_, desc, _ := remote.PopulateImageWithTag("foo")

		desc_fetched, err := tags.Get(ctx, "foo")
		require.NoError(err)
		require.Equal(desc, desc_fetched)

		err = remote.Tags(ctx).Untag(ctx, "foo")
		require.NoError(err)

		desc_loaded, err := tags.Get(ctx, "foo")
		require.NoError(err)
		require.Equal(desc, desc_loaded)
	}))

	t.Run("All", withSvc(func(t *testing.T, local *cache.Repository, remote *registry.Repository, tags distribution.TagService) {
		require := require.New(t)

		remote.PopulateImageWithTag("foo")
		remote.PopulateImageWithTag("bar")

		tags_fetched, err := tags.All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"foo", "bar"}, tags_fetched)

		remote_tags := remote.Tags(ctx)
		err = remote_tags.Untag(ctx, "foo")
		require.NoError(err)
		err = remote_tags.Untag(ctx, "bar")
		require.NoError(err)

		tags_loaded, err := tags.All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"foo", "bar"}, tags_loaded)
	}))
}

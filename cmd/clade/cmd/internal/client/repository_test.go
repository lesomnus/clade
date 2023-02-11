package client_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestRepositoryTags(t *testing.T) {
	ctx := context.Background()
	reg := registry.NewRegistry()
	srv := registry.NewServer(t, reg)

	s := httptest.NewTLSServer(srv.Handler())
	defer s.Close()

	c := client.NewRegistry()
	c.Transport = s.Client().Transport
	c.Cache.Manifests = cache.NewMemManifestCache()
	c.Cache.Tags = cache.NewMemTagCache()

	remote_url, err := url.Parse(s.URL)
	require.NoError(t, err)

	named, err := reference.WithName(remote_url.Host + "/repo/name")
	require.NoError(t, err)

	name := reference.Path(named)
	repo := registry.NewRepository(named)
	repo.PopulateImage()

	tagged, desc, manif := repo.PopulateImage()
	tag := tagged.Tag()

	reg.Repos[name] = repo

	repo_remote, err := c.Repository(named)
	require.NoError(t, err)

	t.Run("manifests", func(t *testing.T) {
		require := require.New(t)

		ms, err := repo_remote.Manifests(ctx)
		require.NoError(err)

		t.Run("by digest", func(t *testing.T) {
			defer c.Cache.Manifests.Clear()

			manif_fetched, err := ms.Get(ctx, desc.Digest)
			require.NoError(err)
			require.Equal(manif, manif_fetched)

			manif_cached, ok := c.Cache.Manifests.GetByDigest(desc.Digest)
			require.True(ok)
			require.Equal(manif, manif_cached)

			dgst := digest.FromString("something")
			_, err = ms.Get(ctx, dgst)
			require.Error(err)

			c.Cache.Manifests.SetByDigest(dgst, manif)
			manif_cached, err = ms.Get(ctx, dgst)
			require.NoError(err)
			require.Equal(manif, manif_cached)
		})

		t.Run("by reference", func(t *testing.T) {
			defer c.Cache.Manifests.Clear()

			manif_fetched, err := ms.Get(ctx, digest.Digest(""), distribution.WithTag(tag))
			require.NoError(err)
			require.Equal(manif, manif_fetched)

			manif_cached, ok := c.Cache.Manifests.GetByRef(tagged)
			require.True(ok)
			require.Equal(manif, manif_cached)

			tagged, err := reference.WithTag(named, "bar")
			require.NoError(err)

			_, err = ms.Get(ctx, digest.Digest(""), distribution.WithTag(tagged.Tag()))
			require.Error(err)

			c.Cache.Manifests.SetByRef(tagged, manif)
			manif_fetched, err = ms.Get(ctx, digest.Digest(""), distribution.WithTag(tagged.Tag()))
			require.NoError(err)
			require.Equal(manif, manif_fetched)
		})
	})

	t.Run("tags", func(t *testing.T) {
		require := require.New(t)

		defer c.Cache.Tags.Clear()

		tags, err := repo.Tags(ctx).All(ctx)
		require.NoError(err)

		tags_fetched, err := repo_remote.Tags(ctx).All(ctx)
		require.NoError(err)
		require.ElementsMatch(tags, tags_fetched)

		tags_cached, ok := c.Cache.Tags.Get(named)
		require.True(ok)
		require.ElementsMatch(tags, tags_cached)

		c.Cache.Tags.Set(named, []string{"bar"})
		tags_cached, err = repo_remote.Tags(ctx).All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"bar"}, tags_cached)
	})

	t.Run("returns ErrorCodeManifestUnknown if tag is not exists", func(t *testing.T) {
		require := require.New(t)

		ms, err := repo_remote.Manifests(ctx)
		require.NoError(err)

		_, err = ms.Get(ctx, digest.Digest(""), distribution.WithTag("not-exists"))
		require.Error(err)

		var errs errcode.Errors
		ok := errors.As(err, &errs)
		require.True(ok)
		require.Len(errs, 1)

		var errc errcode.Error
		ok = errors.As(errs[0], &errc)
		require.True(ok)
		require.Equal(v2.ErrorCodeManifestUnknown.ErrorCode(), errc.Code)
	})
}

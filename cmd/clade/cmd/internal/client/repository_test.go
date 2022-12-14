package client_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
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
	reg := registry.NewRegistry(t)

	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_rul, err := url.Parse(s.URL)
	require.NoError(t, err)

	named, err := reference.ParseNamed(reg_rul.Host + "/repo/name")
	require.NoError(t, err)

	name := reference.Path(named)

	reg.Repos[name] = &registry.Repository{
		Name:      name,
		Manifests: registry.SampleManifests,
	}

	reg_client := client.NewDistRegistry()
	reg_client.Transport = s.Client().Transport
	reg_client.Cache = cache.NewMemCacheStore()

	repo, err := reg_client.Repository(named)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("tags are cached", func(t *testing.T) {
		require := require.New(t)

		defer reg_client.Cache.Clear()

		tags, err := repo.Tags(ctx).All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"foo"}, tags)

		cached_tags, ok := reg_client.Cache.GetTags(named)
		require.True(ok)
		require.ElementsMatch([]string{"foo"}, cached_tags)

		reg_client.Cache.SetTags(named, []string{"bar"})
		cached_tags, err = repo.Tags(ctx).All(ctx)
		require.NoError(err)
		require.ElementsMatch([]string{"bar"}, cached_tags)
	})

	t.Run("get manifest", func(t *testing.T) {
		require := require.New(t)

		svc, err := repo.Manifests(ctx)
		require.NoError(err)

		manifest, err := svc.Get(ctx, digest.Digest(""), distribution.WithTag("foo"))
		require.NoError(err)

		manifest_list, ok := manifest.(*manifestlist.DeserializedManifestList)
		require.True(ok)

		manifests := manifest_list.References()
		require.Len(manifests, 1)
		require.Equal(digest.NewDigestFromEncoded(digest.SHA256, "b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"), manifests[0].Digest)

		manifest_child, err := svc.Get(ctx, manifests[0].Digest)
		require.NoError(err)

		_, ok = manifest_child.(*schema2.DeserializedManifest)
		require.True(ok)
	})

	t.Run("returns ErrorCodeManifestUnknown if tag is not exists", func(t *testing.T) {
		require := require.New(t)

		svc, err := repo.Manifests(ctx)
		require.NoError(err)

		_, err = svc.Get(ctx, digest.Digest(""), distribution.WithTag("not-exists"))
		require.Error(err)

		var errs errcode.Errors
		ok := errors.As(err, &errs)
		require.True(ok)
		require.Len(errs, 1)
		require.ErrorIs(errs[0], v2.ErrorCodeManifestUnknown)
	})
}

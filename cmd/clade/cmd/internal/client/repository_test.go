package client_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
)

func TestRepositoryTags(t *testing.T) {
	require := require.New(t)

	reg := registry.NewRegistry(t)
	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_rul, err := url.Parse(s.URL)
	require.NoError(err)

	named, err := reference.ParseNamed(reg_rul.Host + "/repo/name")
	require.NoError(err)

	name := reference.Path(named)

	reg.Repos[name] = &registry.Repository{
		Name: name,
		Manifests: map[string]registry.Manifest{
			"foo": {},
		},
	}

	reg_client := client.NewDistRegistry()
	reg_client.Transport = s.Client().Transport
	reg_client.Cache = cache.NewMemCacheStore()

	repo, err := reg_client.Repository(named)
	require.NoError(err)

	ctx := context.Background()
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
}

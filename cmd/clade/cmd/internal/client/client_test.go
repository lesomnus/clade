package client_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	require := require.New(t)

	ctx := context.Background()
	reg := registry.NewRegistry()
	srv := registry.NewServer(t, reg)

	s := httptest.NewTLSServer(srv.Handler())
	defer s.Close()

	c := client.NewClient()
	c.Transport = s.Client().Transport

	remote_url, err := url.Parse(s.URL)
	require.NoError(err)

	named, err := reference.WithName(remote_url.Host + "/repo/name")
	require.NoError(err)

	repo := reg.NewRepository(named)
	_, _, manif := repo.PopulateImageWithTag("foo")
	remote_repo, err := c.Repository(named)
	require.NoError(err)

	manifests, err := remote_repo.Manifests(ctx)
	require.NoError(err)

	manif_fetched, err := manifests.Get(ctx, digest.Digest(""), distribution.WithTag("foo"))
	require.NoError(err)
	require.Equal(manif, manif_fetched)
}

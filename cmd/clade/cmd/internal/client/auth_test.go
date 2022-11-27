package client_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/client/auth/challenge"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
)

func TestCredentialStore(t *testing.T) {
	require := require.New(t)

	s := client.NewCredentialStore()

	username, password := s.Basic(nil)
	require.Empty(username)
	require.Empty(password)

	s.SetRefreshToken(nil, "cr.io", "token")
	token := s.RefreshToken(nil, "cr.io")
	require.Equal("token", token)
}

func TestAuthTransport(t *testing.T) {
	require := require.New(t)

	reg := registry.NewRegistry(t)
	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_rul, err := url.Parse(s.URL)
	require.NoError(err)

	named, err := reference.ParseNamed(reg_rul.Host + "/repo/name")
	require.NoError(err)

	res, err := s.Client().Get(s.URL + "/v2/repo/name/tags/list")
	require.NoError(err)
	require.Equal(http.StatusUnauthorized, res.StatusCode)

	tr := &client.AuthTransport{
		Named:            named,
		Base:             s.Client().Transport,
		ChallengeManager: challenge.NewSimpleManager(),
		Credentials:      &client.CredentialStore{},
	}

	s.Client().Transport = tr
	res, err = s.Client().Get(s.URL + "/v2/repo/name/tags/list")

	require.NoError(err)
	require.Equal(http.StatusNotFound, res.StatusCode)
}

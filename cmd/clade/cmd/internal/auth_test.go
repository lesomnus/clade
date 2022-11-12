package internal_test

import (
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/stretchr/testify/require"
)

func TestCredentialStore(t *testing.T) {
	require := require.New(t)

	s := internal.NewCredentialStore()

	username, password := s.Basic(nil)
	require.Empty(username)
	require.Empty(password)

	s.SetRefreshToken(nil, "cr.io", "token")
	token := s.RefreshToken(nil, "cr.io")
	require.Equal("token", token)
}

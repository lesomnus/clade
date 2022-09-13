package internal

import (
	"net/http"
	"net/url"

	"github.com/distribution/distribution/reference"
	"github.com/distribution/distribution/registry/client/auth"
	"github.com/distribution/distribution/registry/client/transport"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"golang.org/x/exp/slices"
)

var (
	DefaultCredentialStore  = &CredentialStore{}
	DefaultChallengeManager = challenge.NewSimpleManager()
)

type CredentialStore struct {
	refresh_tokens map[string]string
}

func (s *CredentialStore) Basic(u *url.URL) (string, string) {
	return "", ""
}

func (s *CredentialStore) RefreshToken(u *url.URL, service string) string {
	return s.refresh_tokens[service]
}

func (s *CredentialStore) SetRefreshToken(u *url.URL, service string, token string) {
	if s.refresh_tokens != nil {
		s.refresh_tokens[service] = token
	}
}

func NewAuthorizer(ref reference.Named, actions ...string) transport.RequestModifier {
	repo := reference.Path(ref)

	if !slices.Contains(actions, "pull") {
		actions = append(actions, "pull")
	}

	return auth.NewAuthorizer(DefaultChallengeManager, auth.NewTokenHandler(http.DefaultTransport, DefaultCredentialStore, repo, actions...))
}

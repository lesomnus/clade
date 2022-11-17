package internal

import (
	"net/url"

	"github.com/docker/distribution/registry/client/auth/challenge"
)

var (
	DefaultCredentialStore  = NewCredentialStore()
	DefaultChallengeManager = challenge.NewSimpleManager()
)

type CredentialStore struct {
	refresh_tokens map[string]string
}

func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		refresh_tokens: make(map[string]string),
	}
}

func (s *CredentialStore) Basic(u *url.URL) (string, string) {
	return "", ""
}

func (s *CredentialStore) RefreshToken(u *url.URL, service string) string {
	return s.refresh_tokens[service]
}

func (s *CredentialStore) SetRefreshToken(u *url.URL, service string, token string) {
	s.refresh_tokens[service] = token
}

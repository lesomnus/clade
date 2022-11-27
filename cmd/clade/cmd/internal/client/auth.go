package client

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/client/auth"
	"github.com/distribution/distribution/v3/registry/client/auth/challenge"
	"github.com/distribution/distribution/v3/registry/client/transport"
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

type AuthTransport struct {
	Named            reference.Named
	Base             http.RoundTripper
	ChallengeManager challenge.Manager
	Credentials      auth.CredentialStore
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	endpoint := url.URL{Scheme: "https", Host: reference.Domain(t.Named), Path: "/v2/"}
	if c, err := t.ChallengeManager.GetChallenges(endpoint); err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	} else if len(c) == 0 {
		if err := func() error {
			client := http.Client{
				Transport: t.Base,
				Timeout:   10 * time.Second,
			}

			res, err := client.Get(endpoint.String())
			if err != nil {
				return err
			}

			defer res.Body.Close()

			if err := t.ChallengeManager.AddResponse(res); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return nil, fmt.Errorf("failed to probe challenge: %w", err)
		}
	}

	authorizer := auth.NewAuthorizer(t.ChallengeManager, auth.NewTokenHandler(t.Base, t.Credentials, reference.Path(t.Named), "pull"))
	return transport.NewTransport(t.Base, authorizer).RoundTrip(req)
}

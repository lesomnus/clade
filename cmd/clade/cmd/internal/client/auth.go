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

type BasicAuth struct {
	Username string
	Password string
}

type CredentialStore struct {
	BasicAuths    map[string]BasicAuth
	RefreshTokens map[string]string
}

func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		BasicAuths:    make(map[string]BasicAuth),
		RefreshTokens: make(map[string]string),
	}
}

func (s *CredentialStore) Basic(u *url.URL) (string, string) {
	if u == nil {
		return "", ""
	}

	if auth, ok := s.BasicAuths[u.Host]; ok {
		return auth.Username, auth.Password
	} else {
		return "", ""
	}
}

func (s *CredentialStore) RefreshToken(u *url.URL, service string) string {
	return s.RefreshTokens[service]
}

func (s *CredentialStore) SetRefreshToken(u *url.URL, service string, token string) {
	s.RefreshTokens[service] = token
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

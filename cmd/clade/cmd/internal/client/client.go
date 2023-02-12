package client

import (
	"net/http"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/client"
	"github.com/distribution/distribution/v3/registry/client/auth/challenge"
)

type Client struct {
	Transport        http.RoundTripper
	ChallengeManager challenge.Manager
	Credentials      *CredentialStore
}

func NewClient() *Client {
	return &Client{
		Transport:        http.DefaultTransport,
		ChallengeManager: challenge.NewSimpleManager(),
		Credentials:      NewCredentialStore(),
	}
}

func (c *Client) Repository(named reference.Named) (distribution.Repository, error) {
	repo_domain := reference.Domain(named)
	repo_path := reference.Path(named)

	tr := &AuthTransport{
		Named:            named,
		Base:             c.Transport,
		ChallengeManager: c.ChallengeManager,
		Credentials:      c.Credentials,
	}

	name_only, err := reference.WithName(repo_path)
	if err != nil {
		return nil, err
	}

	return client.NewRepository(name_only, "https://"+repo_domain, tr)
}

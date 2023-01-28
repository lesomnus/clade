package client

import (
	"net/http"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/client"
	"github.com/distribution/distribution/v3/registry/client/auth/challenge"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
)

// type Registry interface {
// 	Repository(ref reference.Named) (distribution.Repository, error)
// }

type Registry struct {
	Transport        http.RoundTripper
	ChallengeManager challenge.Manager
	Credentials      *CredentialStore
	Cache            cache.CacheStore
}

func NewRegistry() *Registry {
	return &Registry{
		Transport:        http.DefaultTransport,
		ChallengeManager: challenge.NewSimpleManager(),
		Credentials:      NewCredentialStore(),
		Cache:            cache.Cache,
	}
}

func (c *Registry) Repository(named reference.Named) (distribution.Repository, error) {
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

	repo, err := client.NewRepository(name_only, "https://"+repo_domain, tr)
	if err != nil {
		return nil, err
	}

	return &distRepository{
		Repository: repo,
		named:      named,
		cache:      c.Cache,
	}, nil
}

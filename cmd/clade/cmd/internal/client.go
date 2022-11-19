package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/distribution/distribution/reference"
	"github.com/distribution/distribution/registry/client"
	"github.com/distribution/distribution/registry/client/auth"
	"github.com/distribution/distribution/registry/client/transport"
	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/api/errcode"
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/opencontainers/go-digest"
	"golang.org/x/exp/slices"
)

type authTransport struct {
	ref  reference.Named
	base http.RoundTripper
}

var ErrManifestUnknown = errors.New("manifest unknown")

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	endpoint := url.URL{Scheme: "https", Host: reference.Domain(t.ref), Path: "v2/"}
	if c, err := DefaultChallengeManager.GetChallenges(endpoint); err != nil {
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	} else if len(c) == 0 {
		if err := func() error {
			client := http.Client{
				Transport: t.base,
				Timeout:   10 * time.Second,
			}

			res, err := client.Get(endpoint.String())
			if err != nil {
				return err
			}

			defer res.Body.Close()

			if err := DefaultChallengeManager.AddResponse(res); err != nil {
				return err
			}

			return nil
		}(); err != nil {
			return nil, fmt.Errorf("failed to probe challenge: %w", err)
		}
	}

	return t.base.RoundTrip(req)
}

type ClientOption struct {
	transport http.RoundTripper
	actions   []string
}

func NewClientOption() *ClientOption {
	return &ClientOption{
		transport: http.DefaultTransport,
		actions:   make([]string, 0),
	}
}

func (o *ClientOption) apply(modifiers []ClientOptionModifier) {
	for _, m := range modifiers {
		m(o)
	}
}

func WithTransport(transport http.RoundTripper) ClientOptionModifier {
	return func(o *ClientOption) {
		o.transport = transport
	}
}

type ClientOptionModifier func(o *ClientOption)

func NewRepository(ref reference.Named, modifiers ...ClientOptionModifier) (distribution.Repository, error) {
	opt := NewClientOption()
	opt.apply(modifiers)

	if !slices.Contains(opt.actions, "pull") {
		opt.actions = append(opt.actions, "pull")
	}

	repo_path := reference.Path(ref)
	name_only, _ := reference.WithName(repo_path)

	authorizer := auth.NewAuthorizer(DefaultChallengeManager, auth.NewTokenHandler(opt.transport, DefaultCredentialStore, repo_path, opt.actions...))

	tr := &authTransport{
		ref:  ref,
		base: transport.NewTransport(opt.transport, authorizer),
	}

	repo, err := client.NewRepository(name_only, "https://"+reference.Domain(ref), tr)
	if err != nil {
		return nil, err
	}

	return &repoWrapper{
		Repository: repo,
		ref:        ref,
	}, nil
}

type ManifestGetter struct {
	svc distribution.ManifestService
	tag string
	img digest.Digest // Digest for image
}

func NewManifestGetter(ctx context.Context, ref reference.NamedTagged, modifiers ...ClientOptionModifier) (*ManifestGetter, error) {
	repo, err := NewRepository(ref, modifiers...)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository accessor: %w", err)
	}

	img_desc, err := repo.Tags(ctx).Get(ctx, ref.Tag())
	if err != nil {
		var errs errcode.Errors
		if errors.As(err, &errs) {
			for _, e := range errs {
				if errors.Is(e, v2.ErrorCodeManifestUnknown) ||
					errors.Is(e, errcode.ErrorCode(1003)) { // ghcr.io returns 1003 if manifest not exists
					return nil, ErrManifestUnknown
				}
			}
		}

		return nil, fmt.Errorf("failed to get image descriptor: %w", err)
	}

	svc, err := repo.Manifests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest service: %w", err)
	}

	return &ManifestGetter{svc, ref.Tag(), img_desc.Digest}, nil
}

func (g *ManifestGetter) GetByDigest(ctx context.Context, digest digest.Digest) (distribution.Manifest, error) {
	return g.svc.Get(ctx, digest)
}

func (g *ManifestGetter) Get(ctx context.Context) (distribution.Manifest, error) {
	return g.svc.Get(ctx, g.img, distribution.WithTag(g.tag))
}

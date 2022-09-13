package internal

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/distribution/distribution/reference"
	"github.com/distribution/distribution/registry/client"
	"github.com/distribution/distribution/registry/client/transport"
	"github.com/docker/distribution"
	"github.com/docker/distribution/registry/api/errcode"
	v2 "github.com/docker/distribution/registry/api/v2"
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
			res, err := http.Get(endpoint.String())
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

func NewTransport(ref reference.Named, actions ...string) http.RoundTripper {
	return &authTransport{
		ref:  ref,
		base: transport.NewTransport(http.DefaultTransport, NewAuthorizer(ref, actions...)),
	}
}

func NewRepository(ref reference.Named, actions ...string) (distribution.Repository, error) {
	name_only, _ := reference.WithName(reference.Path(ref))
	return client.NewRepository(name_only, "https://"+reference.Domain(ref), NewTransport(ref, actions...))
}

func GetManifest(ctx context.Context, ref reference.NamedTagged) (distribution.Manifest, error) {
	repo, err := NewRepository(ref)
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

	return svc.Get(ctx, img_desc.Digest, distribution.WithTag(ref.Tag()))
}

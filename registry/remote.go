package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1remote "github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// Remote is a Registry backed by go-containerregistry talking to a real
// registry over HTTP(S).
type Remote struct {
	keychain  authn.Keychain
	transport http.RoundTripper
	insecure  bool
}

// RemoteOption configures a Remote.
type RemoteOption func(*Remote)

// WithKeychain sets the authn keychain used to authenticate requests.
// Defaults to authn.DefaultKeychain (the docker config file).
func WithKeychain(k authn.Keychain) RemoteOption {
	return func(r *Remote) { r.keychain = k }
}

// WithTransport sets the HTTP transport. Useful to point at a test registry.
func WithTransport(t http.RoundTripper) RemoteOption {
	return func(r *Remote) { r.transport = t }
}

// WithInsecure parses references against a plain-HTTP registry.
func WithInsecure(insecure bool) RemoteOption {
	return func(r *Remote) { r.insecure = insecure }
}

// NewRemote builds a Remote registry client.
func NewRemote(opts ...RemoteOption) *Remote {
	r := &Remote{keychain: authn.DefaultKeychain}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Remote) nameOpts() []name.Option {
	if r.insecure {
		return []name.Option{name.Insecure}
	}
	return nil
}

func (r *Remote) callOpts(ctx context.Context) []v1remote.Option {
	opts := []v1remote.Option{
		v1remote.WithContext(ctx),
		v1remote.WithAuthFromKeychain(r.keychain),
	}
	if r.transport != nil {
		opts = append(opts, v1remote.WithTransport(r.transport))
	}
	return opts
}

// Tags implements Registry.
func (r *Remote) Tags(ctx context.Context, repo string) ([]string, error) {
	repository, err := name.NewRepository(repo, r.nameOpts()...)
	if err != nil {
		return nil, fmt.Errorf("parse repository %q: %w", repo, err)
	}

	tags, err := v1remote.List(repository, r.callOpts(ctx)...)
	if err != nil {
		return nil, fmt.Errorf("list tags of %q: %w", repo, err)
	}
	return tags, nil
}

// Stat implements Registry.
func (r *Remote) Stat(ctx context.Context, ref string) (*ImageInfo, error) {
	reference, err := name.ParseReference(ref, r.nameOpts()...)
	if err != nil {
		return nil, fmt.Errorf("parse reference %q: %w", ref, err)
	}

	desc, err := v1remote.Get(reference, r.callOpts(ctx)...)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrNotExist
		}
		return nil, fmt.Errorf("get %q: %w", ref, err)
	}

	img, err := desc.Image()
	if err != nil {
		return nil, fmt.Errorf("resolve image %q: %w", ref, err)
	}

	cfg, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("read config of %q: %w", ref, err)
	}

	return &ImageInfo{
		Ref:     ref,
		Digest:  desc.Digest.String(),
		Created: cfg.Created.Time,
		Labels:  cfg.Config.Labels,
	}, nil
}

func isNotFound(err error) bool {
	var terr *transport.Error
	if errors.As(err, &terr) {
		return terr.StatusCode == http.StatusNotFound
	}
	return false
}

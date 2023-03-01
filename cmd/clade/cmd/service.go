package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/ocischema"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
	"github.com/opencontainers/go-digest"
)

var RegistryClient = client.NewClient()
var DefaultCmdService = NewCmdService()

type Namespace interface {
	Repository(named reference.Named) (distribution.Repository, error)
}

type Service interface {
	Output() io.Writer // TODO: Rename to Stdout and add Stderr
	Registry() Namespace

	LoadBuildTreeFromFs(ctx context.Context, bt *clade.BuildTree, path string) error
	LoadBuildGraphFromFs(ctx context.Context, bg *clade.BuildGraph, path string) error
	GetLayer(ctx context.Context, named_tagged reference.NamedTagged) ([]distribution.Descriptor, error)
}

type CmdService struct {
	Sink           io.Writer
	RegistryClient Namespace
}

func NewCmdService() *CmdService {
	return &CmdService{
		Sink:           os.Stdout,
		RegistryClient: cache.WithRemote(resolveRegistryCache(), RegistryClient),
	}
}

func (o *CmdService) Output() io.Writer {
	return o.Sink
}

func (o *CmdService) Registry() Namespace {
	return o.RegistryClient
}

func (o *CmdService) Loader() *load.Loader {
	return &load.Loader{
		Expander: load.Expander{
			Registry: o.Registry(),
		},
	}
}

func (o *CmdService) LoadBuildTreeFromFs(ctx context.Context, bt *clade.BuildTree, path string) error {
	ports, err := load.ReadFromFs(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	return o.Loader().Load(ctx, bt, ports)
}

func (o *CmdService) LoadBuildGraphFromFs(ctx context.Context, bg *clade.BuildGraph, path string) error {
	ports, err := load.ReadFromFs(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	expand := load.Expand{Registry: o.Registry()}
	return expand.Load(ctx, bg, ports)
}

func (s *CmdService) GetLayer(ctx context.Context, named_tagged reference.NamedTagged) ([]distribution.Descriptor, error) {
	repo, err := s.Registry().Repository(named_tagged)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository service: %w", err)
	}

	manif_svc, err := repo.Manifests(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create manifest service: %w", err)
	}

	var manifest distribution.Manifest

	tag := named_tagged.Tag()
	d, err := digest.Parse(tag)
	if err == nil {
		manifest, err = manif_svc.Get(ctx, d)
	} else {
		manifest, err = manif_svc.Get(ctx, digest.Digest(""), distribution.WithTag(tag))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get manifest for tag %s: %w", tag, err)
	}

	return s.getLayer(ctx, manif_svc, manifest)
}

func (o *CmdService) getLayer(ctx context.Context, manif_svc distribution.ManifestService, manifest distribution.Manifest) ([]distribution.Descriptor, error) {
	switch m := manifest.(type) {
	case *ocischema.DeserializedManifest:
		return m.Layers, nil
	case *schema2.DeserializedManifest:
		return m.Layers, nil
	case *manifestlist.DeserializedManifestList:
		if len(m.Manifests) == 0 {
			return nil, errors.New("manifest list is empty")
		}

		// I think it's OK to check only the first one
		// since all the images (with same tag) are updated at once.
		next_manif, err := manif_svc.Get(ctx, m.Manifests[0].Digest)
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest %s: %w", m.Manifests[0].Digest.String(), err)
		}

		return o.getLayer(ctx, manif_svc, next_manif)
	}

	panic("unsupported manifest schema")
}

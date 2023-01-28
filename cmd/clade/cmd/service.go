package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/manifest/manifestlist"
	"github.com/distribution/distribution/v3/manifest/schema2"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
	"github.com/opencontainers/go-digest"
)

var DefaultCmdService = NewCmdService()

type Service interface {
	Output() io.Writer
	LoadBuildTreeFromFs(ctx context.Context, bt *clade.BuildTree, path string) error
	GetLayer(ctx context.Context, named_tagged reference.NamedTagged) ([]distribution.Descriptor, error)
}

type CmdService struct {
	Sink   io.Writer
	Loader load.Loader
}

func NewCmdService() *CmdService {
	return &CmdService{
		Sink:   os.Stdout,
		Loader: load.NewLoader(),
	}
}

func (o *CmdService) Output() io.Writer {
	return o.Sink
}

func (o *CmdService) LoadBuildTreeFromFs(ctx context.Context, bt *clade.BuildTree, path string) error {
	ports, err := load.ReadFromFs(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	return o.Loader.Load(ctx, bt, ports)
}

func (s *CmdService) registry() *client.Registry {
	return s.Loader.Expander.Registry
}

func (s *CmdService) GetLayer(ctx context.Context, named_tagged reference.NamedTagged) ([]distribution.Descriptor, error) {
	repo, err := s.registry().Repository(named_tagged)
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

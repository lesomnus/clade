package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
)

var RegistryClient = client.NewClient()
var DefaultCmdService = NewCmdService()

type Namespace interface {
	Repository(named reference.Named) (distribution.Repository, error)
}

type Service interface {
	Input() io.Reader
	Output() io.Writer

	Registry() Namespace

	LoadBuildGraphFromFs(ctx context.Context, bg *clade.BuildGraph, path string) error
}

type CmdService struct {
	In  io.Reader
	Out io.Writer

	RegistryClient Namespace
}

func NewCmdService() *CmdService {
	return &CmdService{
		In:  os.Stdin,
		Out: os.Stdout,

		RegistryClient: cache.WithRemote(resolveRegistryCache(), RegistryClient),
	}
}

func (o *CmdService) Input() io.Reader {
	return o.In
}

func (o *CmdService) Output() io.Writer {
	return o.Out
}

func (o *CmdService) Registry() Namespace {
	return o.RegistryClient
}

func (o *CmdService) LoadBuildGraphFromFs(ctx context.Context, bg *clade.BuildGraph, path string) error {
	ports, err := clade.ReadPortsFromFs(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	port_loader := clade.PortLoader{Registry: o.Registry()}
	return port_loader.Load(ctx, bg, ports)
}

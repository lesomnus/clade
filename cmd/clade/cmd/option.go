package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
)

type Option struct {
	Output io.Writer
	Loader load.Loader
}

func NewOption() *Option {
	return &Option{
		Output: os.Stdout,
		Loader: load.NewLoader(),
	}
}

func (o *Option) LoadBuildTreeFromFs(ctx context.Context, bt *clade.BuildTree, path string) error {
	ports, err := load.ReadFromFs(path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	return o.Loader.Load(ctx, bt, ports)
}

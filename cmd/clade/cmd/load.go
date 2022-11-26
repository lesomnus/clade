package cmd

import (
	"context"
	"fmt"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
)

func loadBuildTreeFromPorts(ctx context.Context, bt *clade.BuildTree, path string) error {
	ports, err := load.ReadFromFs(path)
	if err != nil {
		return fmt.Errorf("failed to read ports: %w", err)
	}

	loader := load.NewLoader()
	return loader.Load(ctx, bt, ports)
}

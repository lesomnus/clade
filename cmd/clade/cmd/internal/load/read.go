package load

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lesomnus/clade"
	"github.com/rs/zerolog"
)

func ReadFromFs(ctx context.Context, path string) ([]*clade.Port, error) {
	l := zerolog.Ctx(ctx)
	l.Info().Str("path", path).Msg("read ports")

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	ports := make([]*clade.Port, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		port_path := filepath.Join(path, entry.Name(), "port.yaml")
		l.Debug().Str("path", port_path).Msg("read port")

		port, err := clade.ReadPort(port_path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return nil, fmt.Errorf("failed to read port at %s: %w", port_path, err)
		}

		ports = append(ports, port)
	}

	return ports, nil
}

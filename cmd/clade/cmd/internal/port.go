package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lesomnus/clade"
)

func ReadPorts(path string) ([]*clade.Port, error) {
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
		port, err := clade.ReadPort(port_path)
		if err != nil {
			return nil, fmt.Errorf("failed to read port at %s: %w", port_path, err)
		}

		ports = append(ports, port)
	}

	return ports, nil
}

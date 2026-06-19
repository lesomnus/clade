package port

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/goccy/go-yaml"
)

// Filename is the manifest file name expected in a port directory.
const Filename = "port.yaml"

// Load reads and validates the port in the given directory.
func Load(dir string) (*Port, error) {
	p := filepath.Join(dir, Filename)
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p, err)
	}

	var port Port
	if err := yaml.Unmarshal(data, &port); err != nil {
		return nil, fmt.Errorf("decode %s: %w", p, err)
	}

	port.Dir = dir
	if err := port.Validate(); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", p, err)
	}
	return &port, nil
}

// LoadAll scans root for immediate subdirectories that contain a port.yaml and
// loads each one. Subdirectories without a manifest are ignored. The result is
// sorted by directory for deterministic ordering.
func LoadAll(root string) ([]*Port, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read dir %s: %w", root, err)
	}

	ports := make([]*Port, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(root, entry.Name())
		if _, err := os.Stat(filepath.Join(dir, Filename)); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("stat %s: %w", dir, err)
		}

		p, err := Load(dir)
		if err != nil {
			return nil, err
		}
		ports = append(ports, p)
	}

	sort.Slice(ports, func(i, j int) bool { return ports[i].Dir < ports[j].Dir })
	return ports, nil
}

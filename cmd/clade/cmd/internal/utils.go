package internal

import (
	"os"
	"path/filepath"
)

func ResolvePath(base string, path string, fallback string) (string, error) {
	if base == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}

		base = wd
	}

	if !filepath.IsAbs(base) {
		abs, err := filepath.Abs(base)
		if err != nil {
			return "", err
		}

		base = abs
	}

	if path == "" {
		path = fallback
	}

	if filepath.IsAbs(path) {
		return path, nil
	}

	return filepath.Join(base, path), nil
}

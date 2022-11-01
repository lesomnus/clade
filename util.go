package clade

import (
	"os"
	"path/filepath"
)

// ResolvePath returns an absolute representation of joined path of `base` and `path`.
// If the joined path is not absolute, it will be joined with the current working directory to turn it into an absolute path.
// If the `path` is empty, the `base` is joined with `fallback`.
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

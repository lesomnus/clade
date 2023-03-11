package clade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lesomnus/pl"
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

func toString(value any) (string, bool) {
	switch v := value.(type) {
	case interface{ String() string }:
		return v.String(), true
	case string:
		return v, true

	default:
		return "", false
	}
}

func executeBeSingleString(executor *pl.Executor, pl *pl.Pl, data any) (string, error) {
	results, err := executor.Execute(pl, data)
	if err != nil {
		return "", err
	}
	if len(results) != 1 {
		return "", fmt.Errorf("expect result be sized 1 but was %d", len(results))
	}

	v, ok := toString(results[0])
	if !ok {
		return "", fmt.Errorf("expect result be string or stringer")
	}

	return v, nil
}

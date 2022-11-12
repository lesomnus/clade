package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/lesomnus/clade"
)

type rootFlags struct {
	logLevel  string
	portsPath string
}

var root_flags rootFlags

func (f *rootFlags) Evaluate() error {
	switch f.logLevel {
	case "trace":
	case "debug":
	case "info":
	case "warn":
	case "error":
	case "fatal":
	case "":
		f.logLevel = "warn"
	default:
		return fmt.Errorf("invalid log-level value: %s", f.logLevel)
	}

	if p, err := clade.ResolvePath("", f.portsPath, "ports"); err != nil {
		return fmt.Errorf("failed to resolve path to ports: %w", err)
	} else {
		f.portsPath = p
	}

	if fi, err := os.Stat(f.portsPath); err != nil {
		return err
	} else if !fi.IsDir() {
		return errors.New("repository not a directory: " + f.portsPath)
	}

	return nil
}

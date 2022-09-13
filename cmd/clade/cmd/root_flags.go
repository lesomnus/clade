package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
)

type rootFlags struct {
	portsPath string
}

var root_flags rootFlags

func (f *rootFlags) Evaluate() error {
	if p, err := internal.ResolvePath("", f.portsPath, "ports"); err != nil {
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

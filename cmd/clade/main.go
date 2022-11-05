package main

import (
	"os"

	"github.com/lesomnus/clade/cmd/clade/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

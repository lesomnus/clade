package main

import (
	"context"
	"fmt"
	"os"

	"github.com/lesomnus/clade/cmd"
)

func main() {
	c := cmd.NewCmdRoot()
	if err := c.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

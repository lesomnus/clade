package cmd

import (
	"github.com/spf13/cobra"
)

var cache_cmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage caches",
}

func init() {
	root_cmd.AddCommand(cache_cmd)
}

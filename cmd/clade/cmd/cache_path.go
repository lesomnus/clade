package cmd

import (
	"fmt"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/spf13/cobra"
)

var cache_path_cmd = &cobra.Command{
	Use:   "path",
	Short: "Print path cache directory",

	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(internal.Cache.Dir)
		return nil
	},
}

func init() {
	cache_cmd.AddCommand(cache_path_cmd)
}

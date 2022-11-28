package cmd

import (
	"fmt"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/spf13/cobra"
)

func CreateCachePathCmd(svc Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "path",
		Short: "Print path cache directory",

		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(svc.Output(), cache.Cache.Name())
			return nil
		},
	}

	return cmd
}

var (
	cache_path_cmd = CreateCachePathCmd(DefaultCmdService)
)

func init() {
	cache_cmd.AddCommand(cache_path_cmd)
}

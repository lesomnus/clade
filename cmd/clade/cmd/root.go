package cmd

import (
	"github.com/spf13/cobra"
)

var root_cmd = &cobra.Command{
	Use:   "clade",
	Short: "Lade container images with your taste",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return root_flags.Evaluate()
	},
}

func Execute() error {
	return root_cmd.Execute()
}

func init() {
	flags := root_cmd.Flags()
	flags.StringVar(&root_flags.portsPath, "ports", "ports", "Path to repository")
}

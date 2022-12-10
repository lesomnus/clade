package cmd

import "github.com/spf13/cobra"

func CreateChildCmd(flags *TreeFlags, tree_cmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "child Reference",
		Short: "List child images of the given reference",

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags.Strip = 1
			flags.Depth = 1
			flags.Fold = true

			return tree_cmd.RunE(cmd, args)
		},
	}

	cmd_flags := cmd.Flags()
	cmd_flags.BoolVar(&flags.All, "all", false, "Print all images including skipped images")

	return cmd
}

var (
	child_cmd = CreateChildCmd(&tree_flags, tree_cmd)
)

func init() {
	root_cmd.AddCommand(child_cmd)
}

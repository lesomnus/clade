package cmd

import "github.com/spf13/cobra"

func CreateChildCmd(tree_flags *TreeFlags, tree_cmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "child Reference",
		Short: "List child images of the given reference",

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tree_flags.Strip = 1
			tree_flags.Depth = 1
			tree_flags.Fold = true

			return tree_cmd.RunE(cmd, args)
		},
	}

	return cmd
}

var (
	child_cmd = CreateChildCmd(&tree_flags, tree_cmd)
)

func init() {
	root_cmd.AddCommand(child_cmd)
}

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/spf13/cobra"
)

var tree_flags struct {
	strip int
	depth int
}

var tree_cmd = &cobra.Command{
	Use:   "tree [flags] [reference]",
	Short: "List images",

	DisableFlagsInUseLine: true,

	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		bt := make(clade.BuildTree)
		if err := internal.LoadBuildTreeFromPorts(cmd.Context(), bt, root_flags.portsPath); err != nil {
			return fmt.Errorf("failed to load ports at: %w", err)
		}

		var root_node *clade.BuildTreeNode

		if len(args) == 0 {
			root_node = bt.AsNode()
		} else {
			for name, node := range bt {
				if name != args[0] {
					continue
				}

				root_node = &clade.BuildTreeNode{
					Parent:       nil,
					Children:     map[string]*clade.BuildTreeNode{name: node},
					BuildContext: clade.BuildContext{NamedImage: nil},
				}
				break
			}

			if root_node == nil {
				return errors.New(args[0] + " not found")
			}
		}

		root_node.Walk(func(level int, node *clade.BuildTreeNode) error {
			if level < tree_flags.strip {
				return nil
			}

			lv := level - tree_flags.strip

			if tree_flags.depth != 0 && lv >= tree_flags.depth {
				return clade.WalkContinue
			}

			fmt.Print(
				strings.Repeat("\t", lv),
				node.BuildContext.NamedImage.Name.Name(),
				":",
				strings.Join(node.BuildContext.NamedImage.Tags, ", "),
				"\n")
			return nil
		})

		return nil
	},
}

func init() {
	root_cmd.AddCommand(tree_cmd)

	flags := tree_cmd.Flags()
	flags.IntVarP(&tree_flags.strip, "strip", "s", 0, "Skip first n levels")
	flags.IntVarP(&tree_flags.depth, "depth", "d", 0, "Max levels to print")
}

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
	"github.com/spf13/cobra"
)

type TreeFlags struct {
	*RootFlags
	Strip int
	Depth int
	Fold  bool
}

func CreateTreeCmd(flags *TreeFlags, svc Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree [flags] [reference]",
		Short: "List images",

		DisableFlagsInUseLine: true,

		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bt := clade.NewBuildTree()
			if err := svc.LoadBuildTreeFromFs(cmd.Context(), bt, flags.PortsPath); err != nil {
				return fmt.Errorf("failed to load ports: %w", err)
			}

			var root_node *tree.Node[*clade.ResolvedImage]

			if len(args) == 0 {
				root_node = bt.AsNode()
			} else {
				root_name := args[0]

				node, ok := bt.Tree[root_name]
				if !ok {
					return errors.New(root_name + " not found")
				}

				major_name, err := node.Value.Tagged()
				if err != nil {
					panic(fmt.Errorf("broken image def: %s: .tags empty?", root_name))
				}

				node, ok = bt.Tree[major_name.String()]
				if !ok {
					panic(fmt.Errorf("broken build tree: %s: same image with different tag not found", major_name.String()))
				}

				root_node = &tree.Node[*clade.ResolvedImage]{
					Parent:   nil,
					Children: map[string]*tree.Node[*clade.ResolvedImage]{root_name: node},
				}
			}

			visited := make(map[*clade.ResolvedImage]struct{})
			root_node.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
				if level < flags.Strip {
					return nil
				}

				lv := level - flags.Strip

				if flags.Depth != 0 && lv >= flags.Depth {
					return tree.WalkContinue
				}

				if _, ok := visited[node.Value]; !ok {
					visited[node.Value] = struct{}{}
				} else {
					return nil
				}

				image := node.Value
				for _, tag := range image.Tags {
					fmt.Fprint(svc.Output(), strings.Repeat("\t", lv), image.Name(), ":", tag, "\n")

					if flags.Fold {
						break
					}
				}

				return nil
			})

			return nil
		},
	}

	cmd_flags := cmd.Flags()
	cmd_flags.IntVarP(&flags.Strip, "strip", "s", 0, "Skip first n levels")
	cmd_flags.IntVarP(&flags.Depth, "depth", "d", 0, "Max levels to print")
	cmd_flags.BoolVar(&flags.Fold, "fold", false, "Print only primary tags for same images")

	return cmd
}

var (
	tree_flags = TreeFlags{RootFlags: &root_flags}
	tree_cmd   = CreateTreeCmd(&tree_flags, DefaultCmdService)
)

func init() {
	root_cmd.AddCommand(tree_cmd)
}

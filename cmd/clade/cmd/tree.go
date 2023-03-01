package cmd

import (
	"fmt"
	"strings"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/graph"
	"github.com/spf13/cobra"
)

type TreeFlags struct {
	*RootFlags
	All   bool
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
			bg := clade.NewBuildGraph()
			if err := svc.LoadBuildGraphFromFs(cmd.Context(), bg, flags.PortsPath); err != nil {
				return fmt.Errorf("load ports: %w", err)
			}

			root_nodes := bg.Roots()
			if len(args) > 0 {
				node, ok := bg.Get(args[0])
				if !ok {
					return fmt.Errorf(`"%s" not found`, args[0])
				}

				if len(node.Prev) == 0 {
					root_nodes = []*graph.Node[*clade.ResolvedImage]{node}
				} else {
					root_nodes = make([]*graph.Node[*clade.ResolvedImage], 0, len(node.Value.Tags))
					prev, ok := bg.Get(node.Value.From.Primary.String())
					if !ok {
						panic(`prev not exists`)
					}

					for _, sibling := range prev.Next {
						if sibling.Value == node.Value {
							root_nodes = append(root_nodes, sibling)
						}
					}
				}
			}

			var visit func(int, *graph.Node[*clade.ResolvedImage])
			visit = func(level int, node *graph.Node[*clade.ResolvedImage]) {
				effective_next := make([]*graph.Node[*clade.ResolvedImage], 0, len(node.Next))
				for _, next := range node.Next {
					if node.Key() != next.Value.From.Primary.String() {
						continue
					}

					if !flags.All && next.Value.Skip {
						continue
					}

					effective_next = append(effective_next, next)
				}

				if len(node.Next) != 0 && len(effective_next) == 0 {
					return
				}

				if flags.Depth != 0 && level >= flags.Depth {
					return
				}

				if level >= 0 {
					fmt.Fprint(svc.Output(), strings.Repeat("\t", level), node.Key(), "\n")
				}

				for _, next := range effective_next {
					if flags.Fold {
						tagged, err := next.Value.Tagged()
						if err != nil {
							panic("image must have valid tag")
						}
						if next.Key() != tagged.String() {
							continue
						}
					}

					visit(level+1, next)
				}
			}

			for _, node := range root_nodes {
				visit(0-flags.Strip, node)
			}

			return nil
		},
	}

	cmd_flags := cmd.Flags()
	cmd_flags.BoolVar(&flags.All, "all", false, "Print all images including skipped images")
	cmd_flags.IntVarP(&flags.Strip, "strip", "s", 0, "Skip first n levels")
	cmd_flags.IntVarP(&flags.Depth, "depth", "d", 0, "Max levels to print")
	cmd_flags.BoolVar(&flags.Fold, "fold", false, "Print only primary tag for the same images")

	return cmd
}

var (
	tree_flags = TreeFlags{RootFlags: &root_flags}
	tree_cmd   = CreateTreeCmd(&tree_flags, DefaultCmdService)
)

func init() {
	root_cmd.AddCommand(tree_cmd)
}

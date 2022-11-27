package cmd

import (
	"errors"
	"fmt"

	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
	"github.com/spf13/cobra"
)

type OutdatedFlags struct {
	*RootFlags
}

func CreateOutdatedCmd(flags *OutdatedFlags, svc Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "List outdated images",

		RunE: func(cmd *cobra.Command, args []string) error {
			bt := clade.NewBuildTree()
			if err := svc.LoadBuildTreeFromFs(cmd.Context(), bt, flags.PortsPath); err != nil {
				return fmt.Errorf("failed to load ports: %w", err)
			}

			return bt.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
				if level == 0 {
					return nil
				}

				child_name, err := node.Value.Tagged()
				if err != nil {
					return err
				}

				child_layers, err := svc.GetLayer(cmd.Context(), child_name)
				if err != nil {
					var errs errcode.Errors
					if errors.As(err, &errs) {
						if len(errs) == 0 {
							panic("how errors can be empty?")
						}

						for _, err := range errs {
							if errors.Is(err, v2.ErrorCodeManifestUnknown) {
								fmt.Fprintln(svc.Output(), child_name.String())
								return tree.WalkContinue
							}
						}
					}

					return fmt.Errorf("failed to get layers of %s: %w", child_name.String(), err)
				}

				parent_layers, err := svc.GetLayer(cmd.Context(), node.Value.From)
				if err != nil {
					return fmt.Errorf("failed to get layers of %s: %w", node.Value.From.String(), err)
				}

				// repo, err := opt.Registry().Repository(child_name)
				// if err != nil {
				// 	return fmt.Errorf("failed to create repository service: %w", err)
				// }

				// // repo.Manifests(cmd.Context(), distribution.WithTag()

				// child_manif_getter, err := internal.NewManifestGetter(cmd.Context(), child_name)
				// if err != nil {
				// 	if errors.Is(err, internal.ErrManifestUnknown) {
				// 		fmt.Fprintln(o, child_name.String())
				// 		return tree.WalkContinue
				// 	} else {
				// 		return fmt.Errorf("failed to create manifest getter for child image: %w", err)
				// 	}
				// }

				// parent_manif_getter, err := internal.NewManifestGetter(cmd.Context(), node.Value.From)
				// if err != nil {
				// 	return fmt.Errorf("failed to create manifest getter for parent image: %w", err)
				// }

				// child_layers, err := getLayers(cmd.Context(), child_manif_getter)
				// if err != nil {
				// 	return fmt.Errorf("failed to get layers of %s: %w", child_name.String(), err)
				// }

				// parent_layers, err := getLayers(cmd.Context(), parent_manif_getter)
				// if err != nil {
				// 	return fmt.Errorf("failed to get layers of %s: %w", node.Value.From.String(), err)
				// }

				if len(parent_layers) == 0 {
					panic("layer empty")
				}

				is_contains := false
				for _, layer := range child_layers {
					if layer.Digest == parent_layers[len(parent_layers)-1].Digest {
						is_contains = true
						break
					}
				}

				if !is_contains {
					fmt.Fprintln(svc.Output(), child_name.String())
					return tree.WalkContinue
				}

				return nil
			})
		},
	}

	return cmd
}

var (
	outdated_flags = OutdatedFlags{RootFlags: &root_flags}
	outdated_cmd   = CreateOutdatedCmd(&outdated_flags, NewCmdService())
)

func init() {
	root_cmd.AddCommand(outdated_cmd)
}

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

			outdated_images := []string{}
			if err := bt.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
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
								outdated_images = append(outdated_images, child_name.String())
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

				if len(parent_layers) == 0 {
					panic("layer empty")
				}

				is_outdated := false
				for i := 0; i < len(parent_layers); i++ {
					is_outdated = child_layers[i].Digest != parent_layers[i].Digest
					if is_outdated {
						break
					}
				}

				if is_outdated {
					outdated_images = append(outdated_images, child_name.String())
					return tree.WalkContinue
				}

				return nil
			}); err != nil {
				return err
			}

			for _, img := range outdated_images {
				fmt.Fprintln(svc.Output(), img)
			}

			return nil
		},
	}

	return cmd
}

var (
	outdated_flags = OutdatedFlags{RootFlags: &root_flags}
	outdated_cmd   = CreateOutdatedCmd(&outdated_flags, DefaultCmdService)
)

func init() {
	root_cmd.AddCommand(outdated_cmd)
}

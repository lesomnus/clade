package cmd

import (
	"errors"
	"fmt"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/lesomnus/clade/tree"
	"github.com/spf13/cobra"
)

var outdated_cmd = &cobra.Command{
	Use:   "outdated",
	Short: "List outdated images",

	RunE: func(cmd *cobra.Command, args []string) error {
		bt := clade.NewBuildTree()
		if err := internal.LoadBuildTreeFromPorts(cmd.Context(), bt, root_flags.portsPath); err != nil {
			return fmt.Errorf("failed to load ports: %w", err)
		}

		return bt.Walk(func(level int, name string, node *tree.Node[*clade.Image]) error {
			if level == 0 {
				return nil
			}

			child_name, err := node.Value.Tagged()
			if err != nil {
				return err
			}

			child_manifest, err := internal.GetManifest(cmd.Context(), child_name)
			if err != nil {
				if errors.Is(err, internal.ErrManifestUnknown) {
					fmt.Println(child_name.String())
					return tree.WalkContinue
				} else {
					return fmt.Errorf("failed to get manifest for child image: %w", err)
				}
			}

			parent_manifest, err := internal.GetManifest(cmd.Context(), node.Value.From)
			if err != nil {
				return fmt.Errorf("failed to get manifest for parent image: %w", err)
			}

			get_layers := func(manifest distribution.Manifest) []distribution.Descriptor {
				if m, ok := manifest.(*schema2.DeserializedManifest); !ok {
					panic("unsupported manifest schema")
				} else {
					return m.Layers
				}
			}

			child_layers := get_layers(child_manifest)
			parent_layers := get_layers(parent_manifest)

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
				fmt.Println(child_name.String())
				return tree.WalkContinue
			}

			return nil
		})
	},
}

func init() {
	root_cmd.AddCommand(outdated_cmd)
}

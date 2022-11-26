package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/tree"
	"github.com/spf13/cobra"
)

// func getLayers(ctx context.Context, getter *internal.ManifestGetter) ([]distribution.Descriptor, error) {
// 	manif, err := getter.Get(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get manifest: %w", err)
// 	}

// 	switch m := manif.(type) {
// 	case *schema2.DeserializedManifest:
// 		return m.Layers, nil
// 	case *manifestlist.DeserializedManifestList:
// 		if len(m.Manifests) == 0 {
// 			return nil, errors.New("manifest list is empty")
// 		}

// 		// I think it's OK to check only the first one
// 		// since the images are all updated at once.
// 		sub_m, err := getter.GetByDigest(ctx, m.Manifests[0].Digest)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to get manifest: %w", err)
// 		}

// 		switch m := sub_m.(type) {
// 		case *schema2.DeserializedManifest:
// 			return m.Layers, nil
// 		}
// 	}

// 	panic("unsupported manifest schema")
// }

var outdated_cmd = &cobra.Command{
	Use:   "outdated",
	Short: "List outdated images",

	RunE: func(cmd *cobra.Command, args []string) error {
		bt := clade.NewBuildTree()
		if err := loadBuildTreeFromPorts(cmd.Context(), bt, root_flags.portsPath); err != nil {
			return fmt.Errorf("failed to load ports: %w", err)
		}

		o := os.Stdout

		return bt.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
			if level == 0 {
				return nil
			}

			child_name, err := node.Value.Tagged()
			if err != nil {
				return err
			}

			child_manif_getter, err := internal.NewManifestGetter(cmd.Context(), child_name)
			if err != nil {
				if errors.Is(err, internal.ErrManifestUnknown) {
					fmt.Fprintln(o, child_name.String())
					return tree.WalkContinue
				} else {
					return fmt.Errorf("failed to create manifest getter for child image: %w", err)
				}
			}

			parent_manif_getter, err := internal.NewManifestGetter(cmd.Context(), node.Value.From)
			if err != nil {
				return fmt.Errorf("failed to create manifest getter for parent image: %w", err)
			}

			child_layers, err := getLayers(cmd.Context(), child_manif_getter)
			if err != nil {
				return fmt.Errorf("failed to get layers of %s: %w", child_name.String(), err)
			}

			parent_layers, err := getLayers(cmd.Context(), parent_manif_getter)
			if err != nil {
				return fmt.Errorf("failed to get layers of %s: %w", node.Value.From.String(), err)
			}

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
				fmt.Fprintln(o, child_name.String())
				return tree.WalkContinue
			}

			return nil
		})
	},
}

func init() {
	root_cmd.AddCommand(outdated_cmd)
}

package cmd

import (
	"errors"
	"fmt"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/builder"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/spf13/cobra"
)

var build_flags struct {
	dry_run bool
	builder string
}

var build_cmd = &cobra.Command{
	Use:   "build [flags] reference [-- [Builder Args]]",
	Short: "build image",

	DisableFlagsInUseLine: true,

	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		named, err := reference.ParseNamed(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse reference: %w", err)
		}

		builder_args := []string{}
		if len(args) > 1 {
			builder_args = args[1:]
		}

		target_ref, ok := named.(reference.NamedTagged)
		if !ok {
			return errors.New("reference must be tagged")
		}

		b, err := builder.New(build_flags.builder, builder.BuilderConfig{
			DryRun: build_flags.dry_run,
			Args:   builder_args,
		})
		if err != nil {
			return fmt.Errorf("failed to create builder: %w", err)
		}

		bt := clade.NewBuildTree()
		if err := internal.LoadBuildTreeFromPorts(cmd.Context(), bt, root_flags.portsPath); err != nil {
			return fmt.Errorf("failed to load ports: %w", err)
		}

		target_node, ok := bt.Tree[target_ref.String()]
		if !ok {
			return errors.New("failed to find image")
		}

		target_image := target_node.Value
		return b.Build(target_image)
	},
}

func init() {
	root_cmd.AddCommand(build_cmd)

	flags := build_cmd.Flags()
	flags.BoolVar(&build_flags.dry_run, "dry-run", false, "Do not start build")
	flags.StringVar(&build_flags.builder, "builder", "docker-cmd", "Builder to use for the build.")
}

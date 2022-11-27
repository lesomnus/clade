package cmd

import (
	"errors"
	"fmt"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/builder"
	"github.com/spf13/cobra"
)

type BuildFlags struct {
	*RootFlags
	DryRun  bool
	Builder string
}

func CreateBuildCmd(flags *BuildFlags, svc Service) *cobra.Command {
	cmd := &cobra.Command{
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

			b, err := builder.New(flags.Builder, builder.BuilderConfig{
				DryRun: flags.DryRun,
				Args:   builder_args,
			})
			if err != nil {
				return fmt.Errorf("failed to create builder: %w", err)
			}

			bt := clade.NewBuildTree()
			if err := svc.LoadBuildTreeFromFs(cmd.Context(), bt, flags.PortsPath); err != nil {
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

	build_flags := cmd.Flags()
	build_flags.BoolVar(&flags.DryRun, "dry-run", false, "Do not start build")
	build_flags.StringVar(&flags.Builder, "builder", "docker-cmd", "Builder to use for the build.")

	return cmd
}

var (
	build_flags = BuildFlags{RootFlags: &root_flags}
	build_cmd   = CreateBuildCmd(&build_flags, NewCmdService())
)

func init() {
	root_cmd.AddCommand(build_cmd)
}

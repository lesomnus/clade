package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

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
			ctx := cmd.Context()

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

			bg := clade.NewBuildGraph()
			if err := svc.LoadBuildGraphFromFs(ctx, bg, flags.PortsPath); err != nil {
				return fmt.Errorf("failed to load ports: %w", err)
			}

			target_node, ok := bg.Get(target_ref.String())
			if !ok {
				return errors.New("failed to find image")
			}

			target_image := target_node.Value
			dgsts := make([][]byte, 0, 1+len(target_image.From.Secondaries))
			for _, base_image := range target_image.From.All() {
				repo, err := svc.Registry().Repository(base_image)
				if err != nil {
					return fmt.Errorf(`create repository of "%s": %w`, base_image.String(), err)
				}

				desc, err := repo.Tags(ctx).Get(ctx, base_image.Tag())
				if err != nil {
					return fmt.Errorf(`get description of "%s": %w`, base_image.String(), err)
				}

				dgst, err := hex.DecodeString(desc.Digest.Encoded())
				if err != nil {
					return fmt.Errorf(`invalid digest "%s" of "%s": %w`, desc.Digest.String(), base_image.String(), err)
				}

				alias := base_image.Alias
				if alias == "" {
					name := reference.Path(base_image.NamedTagged)
					entries := strings.Split(name, "/")
					if len(entries) == 0 {
						panic("how name can be empty?")
					} else {
						name = entries[len(entries)-1]
					}

					name = strings.ToUpper(name)
					name = strings.ReplaceAll(name, "-", "_")
					alias = name
				}

				target_image.Args[alias] = fmt.Sprintf("%s@%s", base_image.Name(), desc.Digest.String())
				dgsts = append(dgsts, dgst)
			}

			return b.Build(target_image, builder.BuildOption{
				DerefId: clade.CalcDerefId(dgsts...),

				Stdout: svc.Output(),
				Stderr: svc.Output(),
			})
		},
	}

	cmd_flags := cmd.Flags()
	cmd_flags.BoolVar(&flags.DryRun, "dry-run", false, "Do not start build")
	cmd_flags.StringVar(&flags.Builder, "builder", "docker-cmd", "Builder to use for the build.")

	return cmd
}

var (
	build_flags = BuildFlags{RootFlags: &root_flags}
	build_cmd   = CreateBuildCmd(&build_flags, NewCmdService())
)

func init() {
	root_cmd.AddCommand(build_cmd)
}

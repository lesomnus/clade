package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/spf13/cobra"
)

var build_flags struct {
	dry_run bool
}

var build_cmd = &cobra.Command{
	Use:   "build [flags] reference",
	Short: "build image",

	DisableFlagsInUseLine: true,

	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		named, err := reference.ParseNamed(args[0])
		if err != nil {
			return fmt.Errorf("failed to parse reference: %w", err)
		}

		target_ref, ok := named.(reference.NamedTagged)
		if !ok {
			return errors.New("reference must be tagged")
		}

		docker_binary, err := exec.LookPath("docker")
		if err != nil {
			if errors.Is(err, exec.ErrNotFound) {
				if build_flags.dry_run {
					docker_binary = "docker"
				} else {
					return errors.New("docker not found")
				}
			} else {
				return fmt.Errorf("failed to find docker: %w", err)
			}
		}

		bt := make(clade.BuildTree)
		if err := internal.LoadBuildTreeFromPorts(cmd.Context(), bt, root_flags.portsPath); err != nil {
			return fmt.Errorf("failed to load ports at: %w", err)
		}

		target_node, ok := bt[target_ref.String()]
		if !ok {
			return errors.New("failed to find image")
		}

		target_image := target_node.BuildContext.NamedImage
		builder := &exec.Cmd{
			Path:   docker_binary,
			Dir:    target_image.ContextPath,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		}

		builder.Args = func() []string {
			args := []string{builder.Path, "build"}
			args = append(args, "--file", target_image.Dockerfile)

			for _, tag := range target_image.Tags {
				args = append(args, "--tag", fmt.Sprintf("%s:%s", target_image.Name, tag))
			}

			build_args := make(map[string]string)
			for k, v := range target_image.Args {
				build_args[k] = v
			}

			build_args["FROM"] = target_image.From.String()

			for k, v := range build_args {
				args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
			}

			args = append(args, target_image.ContextPath)

			return args
		}()

		fmt.Printf("run: %v\n", builder.Args)

		if build_flags.dry_run {
			return nil
		}

		if err := builder.Run(); err != nil {
			return fmt.Errorf("failed to build: %w", err)
		}

		return nil
	},
}

func init() {
	root_cmd.AddCommand(build_cmd)

	flags := build_cmd.Flags()
	flags.BoolVar(&build_flags.dry_run, "dry-run", false, "Do not start build")
}

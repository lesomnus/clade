package builder

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lesomnus/clade"
	"golang.org/x/exp/slices"
)

type DockerCmdBuilder struct {
	Config BuilderConfig
	binary string

	Platforms []PlatformSpecifier
}

type PlatformSpecifier struct {
	Os   string
	Arch string
}

func (s *PlatformSpecifier) String() string {
	return fmt.Sprintf("%s/%s", s.Os, s.Arch)
}

func (b *DockerCmdBuilder) Build(image *clade.ResolvedImage, option BuildOption) error {
	cmd := &exec.Cmd{
		Path:   b.binary,
		Dir:    image.ContextPath,
		Stdout: option.Stdout,
		Stderr: option.Stderr,
	}

	platforms := make([]string, 0, len(b.Platforms))
	for _, platform := range b.Platforms {
		if !image.Platform.Eval(map[string]bool{
			"t":           true,
			platform.Os:   true,
			platform.Arch: true,
		}) {
			continue
		}

		platforms = append(platforms, platform.String())
	}

	args := []string{b.binary, "buildx", "build"}
	args = append(args, "--file", image.Dockerfile)

	if len(platforms) > 0 {
		args = append(args, "--platform", strings.Join(platforms, ","))
	}

	if option.DerefId != "" {
		if len(platforms) > 1 {
			args = append(args, "--output", fmt.Sprintf("type=image,annotation-index.%s=%s", clade.AnnotationDerefId, option.DerefId))
		} else {
			args = append(args, "--output", fmt.Sprintf("type=image,annotation.%s=%s", clade.AnnotationDerefId, option.DerefId))
		}
	}

	for _, tag := range image.Tags {
		args = append(args, "--tag", fmt.Sprintf("%s:%s", image.Name(), tag))
	}

	for k, v := range image.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, image.ContextPath)
	cmd.Args = args

	if b.Config.DryRun {
		fmt.Fprintln(option.Stdout, cmd.Args)
		return nil
	}

	return cmd.Run()
}

func findDockerBinary() (string, error) {
	bin, err := exec.LookPath("docker")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", errors.New("docker binary not found")
		} else {
			return "", fmt.Errorf("failed to find docker: %w", err)
		}
	}

	return bin, nil
}

func NewDockerCmdBuilder(conf BuilderConfig) (*DockerCmdBuilder, error) {
	bin, err := findDockerBinary()
	if err != nil {
		if conf.DryRun {
			bin = "docker"
		} else {
			return nil, err
		}
	}

	var platforms []PlatformSpecifier
	for {
		i := slices.IndexFunc(conf.Args, func(v string) bool { return v == "--platform" })
		if i < 0 {
			break
		}
		if len(conf.Args) <= i+1 || conf.Args[i+1] == "" || strings.HasPrefix(conf.Args[i+1], "-") {
			return nil, fmt.Errorf("--platform must have value")
		}

		vs := strings.Split(conf.Args[i+1], ",")
		platforms = make([]PlatformSpecifier, 0, len(vs))

		for _, v := range vs {
			specifier := strings.SplitN(v, "/", 2)
			if len(specifier) < 2 || specifier[0] == "" || specifier[1] == "" {
				return nil, fmt.Errorf(`invalid format of "--platform": %s`, v)
			}

			platforms = append(platforms, PlatformSpecifier{
				Os:   specifier[0],
				Arch: specifier[1],
			})
		}

		conf.Args = slices.Delete(conf.Args, i, i+2)
	}

	return &DockerCmdBuilder{
		Config:    conf,
		binary:    bin,
		Platforms: platforms,
	}, nil
}

package builder

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/lesomnus/clade"
)

type dockerBuilder struct {
	config BuilderConfig
	binary string
}

func (b *dockerBuilder) newCmd(image *clade.Image) *exec.Cmd {
	cmd := &exec.Cmd{
		Path:   b.binary,
		Dir:    image.ContextPath,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	args := []string{b.binary, "build"}
	args = append(args, "--file", image.Dockerfile)

	for _, tag := range image.Tags {
		args = append(args, "--tag", fmt.Sprintf("%s:%s", image.Name(), tag))
	}

	args = append(args, "--build-arg", fmt.Sprintf("BASE=%s", image.From.Name()))
	args = append(args, "--build-arg", fmt.Sprintf("TAG=%s", image.From.Tag()))
	for k, v := range image.Args {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, image.ContextPath)
	cmd.Args = args

	return cmd
}

func (b *dockerBuilder) Build(image *clade.Image) error {
	cmd := b.newCmd(image)

	if b.config.DryRun {
		fmt.Println(cmd.Args)
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

func newDockerBuilder(conf BuilderConfig) (*dockerBuilder, error) {
	bin, err := findDockerBinary()
	if err != nil {
		if conf.DryRun {
			bin = "docker"
		} else {
			return nil, err
		}
	}

	return &dockerBuilder{
		config: conf,
		binary: bin,
	}, nil
}

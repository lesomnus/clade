package builder

import (
	"fmt"

	"github.com/lesomnus/clade"
)

type dockerBuildxBuilder struct {
	*dockerBuilder
}

func (b *dockerBuildxBuilder) Build(image *clade.ResolvedImage) error {
	cmd := b.dockerBuilder.newCmd(image)

	args := []string{cmd.Args[0], "buildx", "build"}
	args = append(args, b.config.Args...)
	args = append(args, cmd.Args[2:]...)

	cmd.Args = args

	if b.config.DryRun {
		fmt.Println(cmd.Args)
		return nil
	}

	return cmd.Run()
}

func newDockerBuildxBuilder(conf BuilderConfig) (Builder, error) {
	b, err := newDockerBuilder(conf)
	if err != nil {
		return nil, err
	}

	return &dockerBuildxBuilder{b}, nil
}

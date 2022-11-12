package builder

import (
	"fmt"
	"strings"

	"github.com/lesomnus/clade"
	"golang.org/x/exp/slices"
)

type platformSpecifier struct {
	os   string
	arch string
}

func (s *platformSpecifier) String() string {
	return fmt.Sprintf("%s/%s", s.os, s.arch)
}

type dockerBuildxBuilder struct {
	*dockerBuilder

	platforms []platformSpecifier
}

func (b *dockerBuildxBuilder) Build(image *clade.ResolvedImage) error {
	platforms := make([]string, 0, len(b.platforms))
	for _, platform := range b.platforms {
		data := map[string]bool{
			"t":           true,
			platform.os:   true,
			platform.arch: true,
		}

		if !image.Platform.Eval(data) {
			continue
		}

		platforms = append(platforms, platform.String())
	}

	cmd := b.dockerBuilder.newCmd(image)

	args := []string{cmd.Args[0], "buildx", "build"}

	if len(platforms) > 0 {
		args = append(args, "--platform", strings.Join(platforms, ","))
	}

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

	bx := &dockerBuildxBuilder{
		dockerBuilder: b,
		platforms:     make([]platformSpecifier, 0),
	}

	for {
		i := slices.IndexFunc(b.config.Args, func(v string) bool { return v == "--platform" })
		if i < 0 {
			break
		}
		if len(b.config.Args) <= i+1 {
			return nil, fmt.Errorf("--platform must have value")
		}

		vs := strings.Split(b.config.Args[i+1], ",")
		for _, v := range vs {
			specifier := strings.SplitN(v, "/", 2)
			if len(specifier) < 2 || specifier[0] == "" || specifier[1] == "" {
				return nil, fmt.Errorf("invalid format of --platform value: %s", v)
			}

			bx.platforms = append(bx.platforms, platformSpecifier{
				os:   specifier[0],
				arch: specifier[1],
			})
		}

		b.config.Args = slices.Delete(b.config.Args, i, i+2)
	}

	return bx, nil
}

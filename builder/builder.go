package builder

import (
	"errors"
	"io"

	"github.com/lesomnus/clade"
)

type Builder interface {
	Build(image *clade.ResolvedImage, option BuildOption) error
}

type BuilderConfig struct {
	DryRun bool
	Args   []string
}

type BuildOption struct {
	DerefId string

	Stdout io.Writer
	Stderr io.Writer
}

func New(name string, conf BuilderConfig) (Builder, error) {
	factory, ok := builder_registry[name]
	if !ok {
		return nil, errors.New("not exists")
	}

	return factory(conf)
}

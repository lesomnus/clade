package builder

import (
	"errors"

	"github.com/lesomnus/clade"
)

type Builder interface {
	Build(image *clade.Image) error
}

type BuilderConfig struct {
	DryRun bool
	Args   []string
}

func New(name string, conf BuilderConfig) (Builder, error) {
	factory, ok := builder_registry[name]
	if !ok {
		return nil, errors.New("not exists")
	}

	return factory(conf)
}

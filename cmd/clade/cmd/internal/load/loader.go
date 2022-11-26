package load

import (
	"context"
	"fmt"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/tree"
)

type Loader struct {
	Expander Expander
}

func NewLoader() Loader {
	return Loader{
		Expander: Expander{
			Registry: client.NewDistRegistry(),
		},
	}
}

func (l *Loader) Load(ctx context.Context, bt *clade.BuildTree, ports []*clade.Port) error {
	dt := clade.NewDependencyTree()
	for _, port := range ports {
		for _, image := range port.Images {
			dt.Insert(image)
		}
	}

	return dt.AsNode().Walk(func(level int, name string, node *tree.Node[[]*clade.Image]) error {
		if level == 0 {
			return nil
		}

		for _, image := range node.Value {
			images, err := l.Expander.Expand(ctx, image, bt)
			if err != nil {
				return fmt.Errorf("failed to expand image %s: %w", image.String(), err)
			}

			for _, image := range images {
				if len(image.Tags) == 0 {
					continue
				}

				if err := bt.Insert(image); err != nil {
					return fmt.Errorf("failed to insert image %s into build tree: %w", image.String(), err)
				}
			}
		}

		return nil
	})
}

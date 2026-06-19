package compare

import (
	"context"

	"github.com/lesomnus/clade/registry"
)

func init() {
	Register("created", newCreated)
}

// created compares creation timestamps: the target is outdated if it was
// created before its base image.
type created struct{}

func newCreated(_ []byte) (Comparator, error) {
	return created{}, nil
}

func (created) IsOutdated(_ context.Context, base, target *registry.ImageInfo) (bool, error) {
	return target.Created.Before(base.Created), nil
}

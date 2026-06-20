package compare

import (
	"context"
	"fmt"
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

func (created) IsOutdated(_ context.Context, base, target Comparable) (bool, error) {
	b, ok := base.(Created)
	if !ok {
		return false, fmt.Errorf("created: base: %w", ErrIncomparable)
	}
	t, ok := target.(Created)
	if !ok {
		return false, fmt.Errorf("created: target: %w", ErrIncomparable)
	}
	return t.CreationTime().Before(b.CreationTime()), nil
}

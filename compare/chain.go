package compare

import (
	"context"
	"errors"
)

// ErrIncomparable is returned by a comparator when an operand lacks a capability
// it requires. A Chain treats it as "skip to the next strategy".
var ErrIncomparable = errors.New("compare: operand lacks a required capability")

// ErrNoComparator is returned by a non-empty Chain whose every comparator was
// incomparable. It is a configuration error and aborts the build.
var ErrNoComparator = errors.New("compare: no applicable comparator in chain")

// Spec is a comparator selection: a kind plus its raw YAML params. It decouples
// the per-port compare config and the per-source-kind defaults from the
// constructed Comparators, so neither port nor source needs to import compare's
// concrete types.
type Spec struct {
	Kind   string
	Params []byte
}

// NewChain constructs a Chain from specs, returning an error if any kind is
// unknown.
func NewChain(specs []Spec) (Chain, error) {
	ch := make(Chain, 0, len(specs))
	for _, s := range specs {
		c, err := New(s.Kind, s.Params)
		if err != nil {
			return nil, err
		}
		ch = append(ch, c)
	}
	return ch, nil
}

// Chain tries each comparator in order. The first non-ErrIncomparable result
// wins; any other error aborts. An exhausted non-empty chain returns
// ErrNoComparator. An empty chain is handled by the caller (existence-only) and
// never reaches here.
type Chain []Comparator

func (ch Chain) IsOutdated(ctx context.Context, base, target Comparable) (bool, error) {
	for _, c := range ch {
		outdated, err := c.IsOutdated(ctx, base, target)
		if errors.Is(err, ErrIncomparable) {
			continue
		}
		if err != nil {
			return false, err
		}
		return outdated, nil
	}
	return false, ErrNoComparator
}

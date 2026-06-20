package compare

import (
	"context"
	"errors"
	"testing"
	"time"
)

// capless is a Comparable that implements the seal but no capability, so every
// comparator reports ErrIncomparable against it. The seal means only this
// package can construct such a partial Comparable.
type capless struct{}

func (capless) comparable() {}

// onlyCreated implements just the Created capability.
type onlyCreated struct{ t time.Time }

func (onlyCreated) comparable()               {}
func (c onlyCreated) CreationTime() time.Time { return c.t }

func TestComparatorIncomparable(t *testing.T) {
	ctx := context.Background()

	if _, err := (created{}).IsOutdated(ctx, capless{}, capless{}); !errors.Is(err, ErrIncomparable) {
		t.Errorf("created err = %v, want ErrIncomparable", err)
	}
	if _, err := (digest{}).IsOutdated(ctx, capless{}, capless{}); !errors.Is(err, ErrIncomparable) {
		t.Errorf("digest err = %v, want ErrIncomparable", err)
	}
}

func TestChainFallsBackOnIncomparable(t *testing.T) {
	// digest is incomparable (no Digested/Labeled capability) and the chain
	// falls through to created, which can judge two Created operands.
	ch := Chain{digest{label: DefaultBaseDigestLabel}, created{}}
	base := onlyCreated{t: time.Unix(200, 0)}
	target := onlyCreated{t: time.Unix(100, 0)}

	got, err := ch.IsOutdated(context.Background(), base, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got {
		t.Error("expected outdated via the created fallback")
	}
}

func TestChainExhaustedIsNoComparator(t *testing.T) {
	ch := Chain{created{}, digest{label: DefaultBaseDigestLabel}}
	_, err := ch.IsOutdated(context.Background(), capless{}, capless{})
	if !errors.Is(err, ErrNoComparator) {
		t.Errorf("err = %v, want ErrNoComparator", err)
	}
}

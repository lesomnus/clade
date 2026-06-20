package compare

import (
	"time"

	"github.com/lesomnus/clade/registry"
)

// Comparable is an opaque, sealed view of one existing image. Comparators
// inspect it through the capability interfaces (Created, Digested, Labeled)
// rather than a concrete type. It cannot be implemented outside this package:
// OfImage is the only constructor, which is what makes a comparator's
// capability assertion a reliable contract.
type Comparable interface {
	// comparable seals the interface to this package.
	comparable()
}

// Created exposes an image's creation time.
type Created interface {
	Comparable
	CreationTime() time.Time
}

// Digested exposes an image's own manifest digest. The digest strategy reads it
// from the base to learn the base's current digest.
type Digested interface {
	Comparable
	Digest() string
}

// Labeled exposes an image's config labels by key. The digest strategy reads
// the target's recorded base-digest label through it.
type Labeled interface {
	Comparable
	Label(key string) (string, bool)
}

// imageComparable adapts a registry.ImageInfo. It satisfies every capability;
// comparators still assert so that future, partial Comparables remain valid.
type imageComparable struct{ info *registry.ImageInfo }

func (imageComparable) comparable() {}

func (c imageComparable) CreationTime() time.Time { return c.info.Created }

func (c imageComparable) Digest() string { return c.info.Digest }

func (c imageComparable) Label(key string) (string, bool) {
	v, ok := c.info.Labels[key]
	return v, ok
}

// OfImage wraps registry metadata as a Comparable. A nil info yields a nil
// Comparable.
func OfImage(info *registry.ImageInfo) Comparable {
	if info == nil {
		return nil
	}
	return imageComparable{info: info}
}

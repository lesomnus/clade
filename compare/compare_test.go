package compare_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lesomnus/clade/compare"
	"github.com/lesomnus/clade/registry"
)

func mustNew(t *testing.T, kind, params string) compare.Comparator {
	t.Helper()
	c, err := compare.New(kind, []byte(params))
	if err != nil {
		t.Fatalf("new %q: %v", kind, err)
	}
	return c
}

func at(sec int64) time.Time { return time.Unix(sec, 0) }

func img(info *registry.ImageInfo) compare.Comparable { return compare.OfImage(info) }

func TestCreated(t *testing.T) {
	c := mustNew(t, "created", "")
	base := img(&registry.ImageInfo{Created: at(200)})

	cases := []struct {
		name     string
		target   *registry.ImageInfo
		outdated bool
	}{
		{"older", &registry.ImageInfo{Created: at(100)}, true},
		{"newer", &registry.ImageInfo{Created: at(300)}, false},
		{"equal", &registry.ImageInfo{Created: at(200)}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.IsOutdated(context.Background(), base, img(tc.target))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.outdated {
				t.Errorf("outdated = %v, want %v", got, tc.outdated)
			}
		})
	}
}

func TestDigest(t *testing.T) {
	c := mustNew(t, "digest", "")
	base := img(&registry.ImageInfo{Digest: "sha256:aaa"})
	label := compare.DefaultBaseDigestLabel

	cases := []struct {
		name     string
		target   *registry.ImageInfo
		outdated bool
	}{
		{"match", &registry.ImageInfo{Labels: map[string]string{label: "sha256:aaa"}}, false},
		{"differ", &registry.ImageInfo{Labels: map[string]string{label: "sha256:bbb"}}, true},
		{"missing", &registry.ImageInfo{}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := c.IsOutdated(context.Background(), base, img(tc.target))
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.outdated {
				t.Errorf("outdated = %v, want %v", got, tc.outdated)
			}
		})
	}
}

func TestDigestCustomLabel(t *testing.T) {
	c := mustNew(t, "digest", "kind: digest\nlabel: my.base.digest\n")
	base := img(&registry.ImageInfo{Digest: "sha256:aaa"})
	target := img(&registry.ImageInfo{Labels: map[string]string{"my.base.digest": "sha256:aaa"}})

	got, err := c.IsOutdated(context.Background(), base, target)
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("expected up to date with custom label")
	}
}

func TestUnknownKind(t *testing.T) {
	if _, err := compare.New("nope", nil); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

func TestChainFallback(t *testing.T) {
	// imageComparable satisfies every capability, so a chain returns the first
	// decisive verdict. The ErrIncomparable fallback path against a partial
	// Comparable is covered by the internal test (see incomparable_test.go).
	base := img(&registry.ImageInfo{Created: at(200), Digest: "sha256:aaa"})
	target := img(&registry.ImageInfo{Created: at(100)})

	ch := compare.Chain{mustNew(t, "created", ""), mustNew(t, "digest", "")}
	got, err := ch.IsOutdated(context.Background(), base, target)
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("expected outdated via the created head of the chain")
	}
}

func TestChainEmptyIsNoComparator(t *testing.T) {
	ch := compare.Chain{}
	_, err := ch.IsOutdated(context.Background(), img(&registry.ImageInfo{}), img(&registry.ImageInfo{}))
	if !errors.Is(err, compare.ErrNoComparator) {
		t.Errorf("err = %v, want ErrNoComparator", err)
	}
}

func TestNewChainUnknownKind(t *testing.T) {
	if _, err := compare.NewChain([]compare.Spec{{Kind: "nope"}}); err == nil {
		t.Fatal("expected error for unknown kind in chain")
	}
}

func TestDefaultFor(t *testing.T) {
	if specs := compare.DefaultFor("http"); specs != nil {
		t.Errorf("http default = %v, want nil (existence-only)", specs)
	}
	for _, kind := range []string{"container", ""} {
		specs := compare.DefaultFor(kind)
		if len(specs) != 2 || specs[0].Kind != "created" || specs[1].Kind != "digest" {
			t.Errorf("%q default = %v, want [created digest]", kind, specs)
		}
	}
}

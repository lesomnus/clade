package compare_test

import (
	"context"
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

func TestCreated(t *testing.T) {
	c := mustNew(t, "created", "")
	base := &registry.ImageInfo{Created: at(200)}

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
			got, err := c.IsOutdated(context.Background(), base, tc.target)
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
	base := &registry.ImageInfo{Digest: "sha256:aaa"}
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
			got, err := c.IsOutdated(context.Background(), base, tc.target)
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
	base := &registry.ImageInfo{Digest: "sha256:aaa"}
	target := &registry.ImageInfo{Labels: map[string]string{"my.base.digest": "sha256:aaa"}}

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

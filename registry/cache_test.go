package registry

import (
	"context"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

type countingReg struct {
	inner      Registry
	tags, stat int
}

func (c *countingReg) Tags(ctx context.Context, repo string) ([]string, error) {
	c.tags++
	return c.inner.Tags(ctx, repo)
}

func (c *countingReg) Stat(ctx context.Context, ref string) (*ImageInfo, error) {
	c.stat++
	return c.inner.Stat(ctx, ref)
}

func TestMemCacheExpiry(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	mc := NewMemCache()
	mc.now = clk.now

	mc.Set("k", []byte("v"), time.Minute)
	if _, ok := mc.Get("k"); !ok {
		t.Fatal("expected hit before expiry")
	}

	clk.advance(2 * time.Minute)
	if _, ok := mc.Get("k"); ok {
		t.Fatal("expected miss after expiry")
	}
}

func TestCachedServesFromCache(t *testing.T) {
	fake := NewFake()
	fake.Set("reg.io/x:1", &ImageInfo{Digest: "sha256:a"})

	clk := &fakeClock{t: time.Unix(1000, 0)}
	mc := NewMemCache()
	mc.now = clk.now

	cnt := &countingReg{inner: fake}
	c := &cached{inner: cnt, cache: mc, ttl: time.Minute}

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := c.Stat(ctx, "reg.io/x:1"); err != nil {
			t.Fatal(err)
		}
		if _, err := c.Tags(ctx, "reg.io/x"); err != nil {
			t.Fatal(err)
		}
	}
	if cnt.stat != 1 {
		t.Errorf("stat hit inner %d times, want 1", cnt.stat)
	}
	if cnt.tags != 1 {
		t.Errorf("tags hit inner %d times, want 1", cnt.tags)
	}

	clk.advance(2 * time.Minute)
	if _, err := c.Stat(ctx, "reg.io/x:1"); err != nil {
		t.Fatal(err)
	}
	if cnt.stat != 2 {
		t.Errorf("stat hit inner %d times after expiry, want 2", cnt.stat)
	}
}

func TestCachedDoesNotCacheMissing(t *testing.T) {
	fake := NewFake()
	cnt := &countingReg{inner: fake}
	c := &cached{inner: cnt, cache: NewMemCache(), ttl: time.Minute}

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := c.Stat(ctx, "reg.io/x:missing"); err != ErrNotExist {
			t.Fatalf("err = %v, want ErrNotExist", err)
		}
	}
	if cnt.stat != 2 {
		t.Errorf("missing stat hit inner %d times, want 2 (not cached)", cnt.stat)
	}
}

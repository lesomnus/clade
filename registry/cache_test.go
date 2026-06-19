package registry

import (
	"context"
	"encoding/json"
	"os"
	"sort"
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

func TestFileCacheManagement(t *testing.T) {
	clk := &fakeClock{t: time.Unix(1000, 0)}
	fc, err := NewFileCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fc.now = clk.now

	tags, _ := json.Marshal([]string{"1", "2"})
	fc.Set(KeyTags+"reg.io/x", tags, time.Hour)
	fc.Set(KeyStat+"reg.io/x:1", []byte(`{"digest":"sha256:a"}`), time.Hour)
	fc.Set(KeyTags+"reg.io/y", []byte(`["a"]`), time.Minute)

	entries, err := fc.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}

	byKey := map[string]CacheEntry{}
	for _, e := range entries {
		byKey[e.Key] = e
	}
	if e, ok := byKey[KeyTags+"reg.io/x"]; !ok {
		t.Errorf("missing tag listing entry; keys=%v", keysOf(entries))
	} else if e.Expired {
		t.Error("entry expired before its ttl")
	}

	// y expires after a minute; advance past it.
	clk.advance(2 * time.Minute)
	entries, _ = fc.Entries()
	for _, e := range entries {
		if e.Key == KeyTags+"reg.io/y" && !e.Expired {
			t.Error("expected reg.io/y to be marked expired")
		}
	}

	// Remove reports presence and actually drops the entry.
	if ok, err := fc.Remove(KeyTags + "reg.io/x"); err != nil || !ok {
		t.Fatalf("Remove existing = (%v, %v), want (true, nil)", ok, err)
	}
	if ok, err := fc.Remove(KeyTags + "reg.io/x"); err != nil || ok {
		t.Fatalf("Remove missing = (%v, %v), want (false, nil)", ok, err)
	}

	// Clear wipes whatever remains.
	n, err := fc.Clear()
	if err != nil {
		t.Fatal(err)
	}
	if entries, _ := fc.Entries(); len(entries) != 0 {
		t.Errorf("after Clear, %d entries remain (Clear removed %d)", len(entries), n)
	}
}

func TestFileCacheHealsKeylessEntry(t *testing.T) {
	fc, err := NewFileCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	// Simulate an entry written before keys were stored on disk.
	key := KeyTags + "reg.io/x"
	legacy, _ := json.Marshal(fileEntry{Val: []byte(`["1"]`)})
	if err := os.WriteFile(fc.path(key), legacy, 0o644); err != nil {
		t.Fatal(err)
	}

	// Before a read, the keyless entry is invisible to management.
	if entries, _ := fc.Entries(); len(entries) != 0 {
		t.Fatalf("keyless entry should be skipped, got %v", keysOf(entries))
	}

	// A Get on the hot path heals it in place.
	if _, ok := fc.Get(key); !ok {
		t.Fatal("expected hit on legacy entry")
	}
	entries, _ := fc.Entries()
	if got := keysOf(entries); len(got) != 1 || got[0] != key {
		t.Fatalf("after heal, keys = %v, want [%s]", got, key)
	}
}

func TestFileCacheKeysRecoverable(t *testing.T) {
	fake := NewFake()
	fake.Set("reg.io/x:1", &ImageInfo{Digest: "sha256:a"})

	fc, err := NewFileCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	reg := WithCache(fake, fc, time.Hour)

	ctx := context.Background()
	if _, err := reg.Tags(ctx, "reg.io/x"); err != nil {
		t.Fatal(err)
	}
	if _, err := reg.Stat(ctx, "reg.io/x:1"); err != nil {
		t.Fatal(err)
	}

	entries, err := fc.Entries()
	if err != nil {
		t.Fatal(err)
	}
	keys := keysOf(entries)
	want := []string{KeyStat + "reg.io/x:1", KeyTags + "reg.io/x"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys = %v, want %v", keys, want)
		}
	}
}

func keysOf(entries []CacheEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Key
	}
	sort.Strings(out)
	return out
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

package cache

import (
	"github.com/distribution/distribution/reference"
	"github.com/distribution/distribution/v3"
	"github.com/opencontainers/go-digest"
)

type ManifestCache interface {
	Name() string
	Clear() error
	GetByDigest(digest digest.Digest) (distribution.Manifest, bool)
	SetByDigest(digest digest.Digest, manifest distribution.Manifest)
	GetByRef(named reference.NamedTagged) (distribution.Manifest, bool)
	SetByRef(named reference.NamedTagged, manifest distribution.Manifest)
}

type NullManifestCache struct{}

func (c *NullManifestCache) Name() string {
	return "@null"
}

func (c *NullManifestCache) Clear() error {
	return nil
}

func (c *NullManifestCache) GetByDigest(digest digest.Digest) (distribution.Manifest, bool) {
	return nil, false
}

func (c *NullManifestCache) SetByDigest(digest digest.Digest, manifest distribution.Manifest) {}

func (c *NullManifestCache) GetByRef(named reference.Named) (distribution.Manifest, bool) {
	return nil, false
}

func (c *NullManifestCache) SetByRef(named reference.Named, manifest distribution.Manifest) {}

type MemManifestCache struct {
	ByDigest map[string]distribution.Manifest
	ByRef    map[string]distribution.Manifest
}

func NewMemManifestCache() *MemManifestCache {
	return &MemManifestCache{
		ByDigest: make(map[string]distribution.Manifest),
		ByRef:    make(map[string]distribution.Manifest),
	}
}

func (c *MemManifestCache) Name() string {
	return "@mem"
}

func (c *MemManifestCache) Clear() error {
	*c = *NewMemManifestCache()
	return nil
}

func (c *MemManifestCache) GetByDigest(digest digest.Digest) (distribution.Manifest, bool) {
	manif, ok := c.ByDigest[digest.String()]
	return manif, ok
}

func (c *MemManifestCache) SetByDigest(digest digest.Digest, manifest distribution.Manifest) {
	c.ByDigest[digest.String()] = manifest
}

func (c *MemManifestCache) GetByRef(named reference.NamedTagged) (distribution.Manifest, bool) {
	manif, ok := c.ByRef[named.String()]
	return manif, ok
}

func (c *MemManifestCache) SetByRef(named reference.NamedTagged, manifest distribution.Manifest) {
	c.ByRef[named.String()] = manifest
}

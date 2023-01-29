package cache

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/distribution/distribution/reference"
	"github.com/distribution/distribution/v3"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/opencontainers/go-digest"
)

type ManifestCache interface {
	Name() string
	Clear() error
	GetByDigest(dgst digest.Digest) (distribution.Manifest, bool)
	SetByDigest(dgst digest.Digest, manifest distribution.Manifest)
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

func (c *NullManifestCache) GetByDigest(dgst digest.Digest) (distribution.Manifest, bool) {
	return nil, false
}

func (c *NullManifestCache) SetByDigest(dgst digest.Digest, manifest distribution.Manifest) {}

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

func (c *MemManifestCache) GetByDigest(dgst digest.Digest) (distribution.Manifest, bool) {
	manif, ok := c.ByDigest[dgst.String()]
	return manif, ok
}

func (c *MemManifestCache) SetByDigest(dgst digest.Digest, manifest distribution.Manifest) {
	c.ByDigest[dgst.String()] = manifest
}

func (c *MemManifestCache) GetByRef(named reference.NamedTagged) (distribution.Manifest, bool) {
	manif, ok := c.ByRef[named.String()]
	return manif, ok
}

func (c *MemManifestCache) SetByRef(named reference.NamedTagged, manifest distribution.Manifest) {
	c.ByRef[named.String()] = manifest
}

type FsManifestCache struct {
	fsCache
	*MemManifestCache
}

func NewFsManifestCache(dir string) *FsManifestCache {
	return &FsManifestCache{
		fsCache:          fsCache{Dir: dir},
		MemManifestCache: NewMemManifestCache(),
	}
}

func (c *FsManifestCache) Name() string {
	return c.fsCache.Name()
}

func (c *FsManifestCache) Clear() error {
	c.MemManifestCache.Clear()
	return c.fsCache.Clear()
}

func (c *FsManifestCache) ToPath(dgst digest.Digest) string {
	return filepath.Join(c.Dir, dgst.Algorithm().String(), dgst.Encoded())
}

func (c *FsManifestCache) GetByDigest(dgst digest.Digest) (distribution.Manifest, bool) {
	tgt := c.ToPath(dgst)

	data, err := os.ReadFile(tgt)
	if err != nil {
		return nil, false
	}

	sep := bytes.IndexRune(data, '\n')

	manif, _, err := distribution.UnmarshalManifest(string(data[:sep]), data[sep+1:])
	if err != nil {
		return nil, false
	}

	return manif, true
}

func (c *FsManifestCache) SetByDigest(dgst digest.Digest, manifest distribution.Manifest) {
	tgt := c.ToPath(dgst)
	if err := os.MkdirAll(filepath.Dir(tgt), 0755); err != nil {
		return
	}

	media_type, data, err := manifest.Payload()
	if err != nil {
		return
	}

	f, err := os.OpenFile(tgt, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}

	defer f.Close()

	f.WriteString(media_type)
	f.WriteString("\n")
	f.Write(data)
}

func init() {
	distribution.RegisterManifestSchema("testing", func(data []byte) (distribution.Manifest, distribution.Descriptor, error) {
		return &registry.Manifest{ContentType: "testing", Blob: data}, distribution.Descriptor{}, nil
	})
}

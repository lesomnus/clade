package registry

import (
	"context"
	"io"
	"net/http"

	"github.com/distribution/distribution/v3"
	"github.com/opencontainers/go-digest"
)

type BlobStore struct {
	Repo *Repository
}

func (s *BlobStore) Stat(ctx context.Context, dgst digest.Digest) (distribution.Descriptor, error) {
	panic("not implemented")
}

func (s *BlobStore) Get(ctx context.Context, dgst digest.Digest) ([]byte, error) {
	panic("not implemented")
}

func (s *BlobStore) Open(ctx context.Context, dgst digest.Digest) (io.ReadSeekCloser, error) {
	panic("not implemented")
}

func (s *BlobStore) Put(ctx context.Context, mediaType string, p []byte) (distribution.Descriptor, error) {
	panic("not implemented")
}

func (s *BlobStore) Create(ctx context.Context, options ...distribution.BlobCreateOption) (distribution.BlobWriter, error) {
	panic("not implemented")
}

func (s *BlobStore) Resume(ctx context.Context, id string) (distribution.BlobWriter, error) {
	panic("not implemented")
}

func (s *BlobStore) ServeBlob(ctx context.Context, w http.ResponseWriter, r *http.Request, dgst digest.Digest) error {
	panic("not implemented")
}

func (s *BlobStore) Delete(ctx context.Context, dgst digest.Digest) error {
	panic("not implemented")
}

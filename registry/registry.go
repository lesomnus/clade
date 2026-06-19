// Package registry abstracts read access to a container image registry behind
// a narrow interface so that the concrete client (go-containerregistry), the
// metadata cache and the test fake are interchangeable.
//
// Querying image metadata consumes registry rate limits, so callers are
// expected to wrap a Registry with a cache (see WithCache).
package registry

import (
	"context"
	"errors"
	"time"
)

// ErrNotExist is returned by Stat when the referenced image does not exist.
var ErrNotExist = errors.New("image does not exist")

// ImageInfo is the subset of image metadata clade needs to build the graph and
// decide whether a target is outdated.
type ImageInfo struct {
	// Ref is the reference the info was fetched for, "repo:tag".
	Ref string `json:"ref"`
	// Digest is the manifest digest, "sha256:...".
	Digest string `json:"digest"`
	// Created is the image creation time from its config.
	Created time.Time `json:"created"`
	// Labels are the image config labels.
	Labels map[string]string `json:"labels,omitempty"`
}

// Registry provides read-only access to image metadata.
type Registry interface {
	// Tags lists the tags of a repository, e.g. "docker.io/library/golang".
	Tags(ctx context.Context, repo string) ([]string, error)
	// Stat returns metadata for a reference, "repo:tag". It returns
	// ErrNotExist if the image is absent.
	Stat(ctx context.Context, ref string) (*ImageInfo, error)
}

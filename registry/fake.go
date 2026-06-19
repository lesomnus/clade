package registry

import (
	"context"
	"sort"
	"strings"
	"sync"
)

// Fake is an in-memory Registry for tests. References are treated as opaque
// "repo:tag" strings without normalization, matching how the graph builder
// composes references.
type Fake struct {
	mu    sync.Mutex
	repos map[string]map[string]*ImageInfo // repo -> tag -> info
}

// NewFake creates an empty fake registry.
func NewFake() *Fake {
	return &Fake{repos: map[string]map[string]*ImageInfo{}}
}

// Set adds or replaces the image at ref ("repo:tag").
func (f *Fake) Set(ref string, info *ImageInfo) {
	repo, tag := splitRef(ref)

	f.mu.Lock()
	defer f.mu.Unlock()

	tags, ok := f.repos[repo]
	if !ok {
		tags = map[string]*ImageInfo{}
		f.repos[repo] = tags
	}

	clone := *info
	clone.Ref = ref
	tags[tag] = &clone
}

// Tags implements Registry.
func (f *Fake) Tags(_ context.Context, repo string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	tags := make([]string, 0, len(f.repos[repo]))
	for tag := range f.repos[repo] {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

// Stat implements Registry.
func (f *Fake) Stat(_ context.Context, ref string) (*ImageInfo, error) {
	repo, tag := splitRef(ref)

	f.mu.Lock()
	defer f.mu.Unlock()

	info, ok := f.repos[repo][tag]
	if !ok {
		return nil, ErrNotExist
	}

	clone := *info
	return &clone, nil
}

// splitRef splits "repo:tag" into its parts. The tag is the substring after the
// last colon that follows the last slash, so registry ports (e.g.
// "localhost:5000/x:tag") are handled correctly. Returns an empty tag if none.
func splitRef(ref string) (repo, tag string) {
	colon := strings.LastIndex(ref, ":")
	slash := strings.LastIndex(ref, "/")
	if colon < 0 || colon < slash {
		return ref, ""
	}
	return ref[:colon], ref[colon+1:]
}

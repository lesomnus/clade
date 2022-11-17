package internal_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"golang.org/x/exp/slices"
)

// Implements mock container registry for testing.

type repository struct {
	tags []string
}

type registry struct {
	t *testing.T

	repos map[string]repository
}

func (r *registry) handleRoot(res http.ResponseWriter, req *http.Request) {
	r.t.Logf("[%5s] %s", req.Method, req.URL.String())

	// req.

	if !strings.HasPrefix(req.URL.Path, "/v2/") {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	if req.URL.Path == "/v2/" {
		res.WriteHeader(http.StatusOK)
		return
	}

	parts := strings.Split(req.URL.Path, "/")
	parts = parts[2:]

	if len(parts) < 3 {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	name := strings.Join(parts[0:2], "/")
	path := strings.Join(parts[2:], "/")
	switch path {
	case "tags/list":
		r.handleTagsAll(name, res, req)

	default:
		res.WriteHeader(http.StatusNotFound)
		return
	}
}

func (r *registry) handleTagsAll(name string, res http.ResponseWriter, req *http.Request) {
	repo, ok := r.repos[name]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	tags := slices.Clone(repo.tags)
	for i, tag := range tags {
		tags[i] = fmt.Sprintf(`"%s"`, tag)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf(`{
	"name": "%s",
	"tags": [%s]
}`,
		name,
		strings.Join(tags, ","))))
}

func (r *registry) handler() http.HandlerFunc {
	return http.HandlerFunc(r.handleRoot)
}

func newRegistry(t *testing.T) *registry {
	return &registry{
		t:     t,
		repos: make(map[string]repository),
	}
}

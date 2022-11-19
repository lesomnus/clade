package internal_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"golang.org/x/exp/slices"
)

// Implements mock container registry for testing.

type manifest struct {
	contentType string
	blob        []byte
}

type repository struct {
	name      string
	manifests map[string]manifest
}

type registry struct {
	t *testing.T

	repos map[string]*repository
}

func (r *repository) Tags() []string {
	tags := make([]string, len(r.manifests))

	i := 0
	for tag := range r.manifests {
		tags[i] = tag
		i++
	}

	return tags
}

type registryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *registry) handleRoot(res http.ResponseWriter, req *http.Request) {
	r.t.Logf("[%5s] %s", req.Method, req.URL.String())

	if req.URL.Path == "/" && req.URL.Query().Has("scope") {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(fmt.Sprintf(`{"token": "%s", "expires_in": 3600, "issued_at": "%s"}`, "t0ken", time.Now().Format(time.RFC3339))))
		return
	}

	if req.Header.Get("Authorization") == "" {
		res.Header().Add("WWW-Authenticate", fmt.Sprintf(`Bearer realm="https://%s"`, req.Host))
		res.WriteHeader(http.StatusUnauthorized)
		res.Write([]byte(`{"errors": [{"code": "UNAUTHORIZED", "message": "authentication required", "detail": null}]}`))
		return
	}

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
	parts = append(parts[2:], "")

	repo, ok := r.repos[name]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	var errs []registryError = nil
	switch parts[0] {
	case "tags":
		switch parts[1] {
		case "list":
			errs = r.handleTagsList(res, req, repo)
		}

	case "manifests":
		if parts[1] == "" {
			break
		}

		errs = r.handleManifests(res, req, repo, parts[1])
	}

	if errs == nil {
		res.WriteHeader(http.StatusNotFound)
		errs = []registryError{{
			Code:    "UNSUPPORTED",
			Message: "unknown endpoint",
		}}
	} else if len(errs) == 0 {
		return
	}

	err_res, err := json.Marshal(errs)
	if err != nil {
		r.t.Fatal(err)
	}

	res.Write([]byte(fmt.Sprintf(`{"errors":%s}`, string(err_res))))
}

func (r *registry) handleTagsList(res http.ResponseWriter, req *http.Request, repo *repository) []registryError {
	tags := slices.Clone(repo.Tags())
	for i, tag := range tags {
		tags[i] = fmt.Sprintf(`"%s"`, tag)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf(`{"name": "%s", "tags": [%s]}`, repo.name, strings.Join(tags, ","))))

	return []registryError{}
}

func (r *registry) handleManifests(res http.ResponseWriter, req *http.Request, repo *repository, tag string) []registryError {
	manifest, ok := repo.manifests[tag]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return []registryError{{Code: "MANIFEST_UNKNOWN", Message: "manifest unknown"}}
	}

	res.Header().Set("Content-Type", manifest.contentType)
	res.Header().Set("Content-Length", fmt.Sprint(len(manifest.blob)))
	res.WriteHeader(http.StatusOK)

	if req.Method == "HEAD" {
		return []registryError{}
	}

	io.Copy(res, bytes.NewReader(manifest.blob))

	return []registryError{}
}

func (r *registry) handler() http.HandlerFunc {
	return http.HandlerFunc(r.handleRoot)
}

func newRegistry(t *testing.T) *registry {
	return &registry{
		t:     t,
		repos: make(map[string]*repository),
	}
}

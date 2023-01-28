package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/distribution/distribution/v3"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

// Implements mock container registry for testing.

type Manifest struct {
	ContentType string
	Blob        []byte
}

func (m *Manifest) References() []distribution.Descriptor {
	return []distribution.Descriptor{}
}

func (m *Manifest) Payload() (string, []byte, error) {
	return m.ContentType, m.Blob, nil
}

type Repository struct {
	Name      string
	Manifests map[string]Manifest
}

type Registry struct {
	T *testing.T

	Repos      map[string]*Repository
	EnableAuth bool
}

func NewRegistry(t *testing.T) *Registry {
	return &Registry{
		T: t,

		Repos:      make(map[string]*Repository),
		EnableAuth: true,
	}
}

func (r *Repository) Tags() []string {
	tags := make([]string, 0, len(r.Manifests))

	for tag := range r.Manifests {
		if strings.HasPrefix(tag, "sha256:") {
			continue
		}

		tags = append(tags, tag)
	}

	return tags
}

type registryError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *Registry) handleRoot(res http.ResponseWriter, req *http.Request) {
	r.T.Logf("[%5s] %s", req.Method, req.URL.String())

	if req.URL.Path == "/" && req.URL.Query().Has("scope") {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(fmt.Sprintf(`{"token": "%s", "expires_in": 3600, "issued_at": "%s"}`, "t0ken", time.Now().Format(time.RFC3339))))
		return
	}

	if r.EnableAuth && req.Header.Get("Authorization") == "" {
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

	repo, ok := r.Repos[name]
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
	require.NoError(r.T, err)

	res.Write([]byte(fmt.Sprintf(`{"errors":%s}`, string(err_res))))
}

func (r *Registry) handleTagsList(res http.ResponseWriter, req *http.Request, repo *Repository) []registryError {
	tags := slices.Clone(repo.Tags())
	for i, tag := range tags {
		tags[i] = fmt.Sprintf(`"%s"`, tag)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf(`{"name": "%s", "tags": [%s]}`, repo.Name, strings.Join(tags, ","))))

	return []registryError{}
}

func (r *Registry) handleManifests(res http.ResponseWriter, req *http.Request, repo *Repository, tag string) []registryError {
	manifest, ok := repo.Manifests[tag]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return []registryError{{Code: "MANIFEST_UNKNOWN", Message: "manifest unknown"}}
	}

	res.Header().Set("Content-Type", manifest.ContentType)
	res.Header().Set("Content-Length", fmt.Sprint(len(manifest.Blob)))
	res.WriteHeader(http.StatusOK)

	if req.Method == "HEAD" {
		return []registryError{}
	}

	io.Copy(res, bytes.NewReader(manifest.Blob))

	return []registryError{}
}

func (r *Registry) Handler() http.HandlerFunc {
	return http.HandlerFunc(r.handleRoot)
}

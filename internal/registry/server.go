package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

type Server struct {
	*Registry

	T *testing.T

	EnableAuth bool
}

func NewServer(t *testing.T, reg *Registry) *Server {
	return &Server{
		Registry: reg,

		T: t,

		EnableAuth: true,
	}
}

func (s *Server) handleRoot(res http.ResponseWriter, req *http.Request) {
	require := require.New(s.T)

	s.T.Logf("[%5s] %s", req.Method, req.URL.String())

	if req.URL.Path == "/" && req.URL.Query().Has("scope") {
		res.WriteHeader(http.StatusOK)
		res.Write([]byte(fmt.Sprintf(`{"token": "%s", "expires_in": 3600, "issued_at": "%s"}`, "very_secure", time.Now().Format(time.RFC3339))))
		return
	}

	if s.EnableAuth && req.Header.Get("Authorization") == "" {
		ecs := NewErrorCodes(ErrCodeUnauthorized)
		data, err := json.Marshal(ecs)
		require.NoError(err)

		res.Header().Add("WWW-Authenticate", fmt.Sprintf(`Bearer realm="https://%s"`, req.Host))
		res.WriteHeader(http.StatusUnauthorized)
		res.Write(data)
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

	named, err := reference.ParseNamed(fmt.Sprint(req.Host, "/", name))
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}

	repo, ok := s.Repos[named.Name()]
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	var (
		done        bool
		unsupported bool
	)
	switch parts[0] {
	case "manifests":
		if parts[1] == "" {
			unsupported = true
			break
		}

		done, err = s.handleManifests(res, req, repo, parts[1])

	case "tags":
		switch parts[1] {
		case "list":
			done, err = s.handleTagsList(res, req, repo)
		}

	default:
		unsupported = true
	}

	if unsupported {
		data, err := json.Marshal(NewErrorCodes(ErrCodeUnsupported))
		require.NoError(err)

		res.WriteHeader(http.StatusNotFound)
		res.Write(data)
		return
	}

	if done {
		return
	}

	if err == nil {
		res.WriteHeader(http.StatusOK)
		return
	}

	ecs, ok := err.(*errcode.Errors)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(ecs)
	require.NoError(err)

	res.Write(data)
}

func (s *Server) handleManifests(res http.ResponseWriter, req *http.Request, repo *Repository, ref string) (bool, error) {
	ctx := req.Context()
	ms, err := repo.Manifests(ctx)
	if err != nil {
		return false, err
	}

	dgst := digest.Digest("")
	if strings.Contains(ref, ":") {
		dgst, err = digest.Parse(ref)
		if err != nil {
			if errors.Is(err, digest.ErrDigestUnsupported) {
				res.WriteHeader(http.StatusNotImplemented)
				return true, nil
			}
			if errors.Is(err, digest.ErrDigestInvalidLength) || errors.Is(err, digest.ErrDigestInvalidFormat) {
				res.WriteHeader(http.StatusBadRequest)
				return true, nil
			}

			return false, err
		}
	} else {
		desc, err := repo.Tags(ctx).Get(ctx, ref)
		if err == nil {
			dgst = desc.Digest
		} else {
			if errs, ok := err.(errcode.Errors); ok {
				for _, err := range errs {
					if errors.Is(err, v2.ErrorCodeManifestUnknown) {
						res.WriteHeader(http.StatusNotFound)
						return false, NewErrorCodes(ErrCodeManifestUnknown)
					}
				}
			}

			return false, err
		}
	}

	manif, err := ms.Get(ctx, dgst)
	if err != nil {
		if errs, ok := err.(errcode.Errors); ok {
			for _, err := range errs {
				if errors.Is(err, v2.ErrorCodeManifestUnknown) {
					res.WriteHeader(http.StatusNotFound)
					return false, NewErrorCodes(ErrCodeManifestUnknown)
				}
			}
		}

		return false, err
	}

	mt, data, err := manif.Payload()
	if err != nil {
		return false, err
	}

	res.Header().Set("Content-Type", mt)
	res.Header().Set("Content-Length", fmt.Sprint(len(data)))
	res.Header().Set("Docker-Content-Digest", dgst.String())
	res.WriteHeader(http.StatusOK)

	if req.Method == http.MethodGet {
		io.Copy(res, bytes.NewReader(data))
	}

	return true, nil
}

func (s *Server) handleTagsList(res http.ResponseWriter, req *http.Request, repo *Repository) (bool, error) {
	ctx := req.Context()
	tags, err := repo.Tags(ctx).All(ctx)
	if err != nil {
		if errs, ok := err.(errcode.Errors); ok {
			for _, err := range errs {
				if errors.Is(err, v2.ErrorCodeManifestUnknown) {
					res.WriteHeader(http.StatusNotFound)
					return false, NewErrorCodes(ErrCodeManifestUnknown)
				}
			}
		}

		return false, err
	}

	for i, tag := range tags {
		tags[i] = fmt.Sprintf(`"%s"`, tag)
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(fmt.Sprintf(`{"name": "%s", "tags": [%s]}`, repo.named.Name(), strings.Join(tags, ","))))
	return true, nil
}

func (s *Server) Handler() http.HandlerFunc {
	return http.HandlerFunc(s.handleRoot)
}

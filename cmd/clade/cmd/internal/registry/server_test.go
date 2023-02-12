package registry_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
)

type AuthTransport struct {
	base http.RoundTripper
}

func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "foo")

	return t.base.RoundTrip(req)
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	reg := registry.NewRegistry()
	srv := registry.NewServer(t, reg)

	named, err := reference.WithName("repo/name")
	require.NoError(t, err)

	repo := reg.NewRepository(named)
	repo.PopulateImage()

	tagged, desc, _ := repo.PopulateImage()
	tag := tagged.Tag()

	s := httptest.NewServer(srv.Handler())
	defer s.Close()

	t.Run("401 if no authenticate header", func(t *testing.T) {
		require := require.New(t)

		paths := []string{"/v2", "/v2/tags", "/v2/not-exists"}
		for _, p := range paths {
			res, err := s.Client().Get(s.URL + p)
			require.NoError(err)
			require.Equal(http.StatusUnauthorized, res.StatusCode)
		}
	})

	t.Run("/ returns token", func(t *testing.T) {
		require := require.New(t)

		res, err := s.Client().Get(s.URL + "?scope=somewhere")
		require.NoError(err)
		require.Equal(http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(err)
		require.Contains(string(body), "token")
	})

	t.Run("/v2/ returns OK if request has authenticate header", func(t *testing.T) {
		require := require.New(t)

		req, err := http.NewRequest(http.MethodGet, s.URL+"/v2/", nil)
		require.NoError(err)

		req.Header.Set("Authorization", "foo")

		res, err := s.Client().Do(req)
		require.NoError(err)
		require.Equal(http.StatusOK, res.StatusCode)
	})

	// Following tests uses authenticated requests.
	s.Client().Transport = &AuthTransport{s.Client().Transport}

	t.Run("404 if page not exists", func(t *testing.T) {
		require := require.New(t)

		paths := []string{
			"/not-exists",
			"/v1",
			"/v2/not-exists",
			"/v2/repo/not-exists",
			"/v2/repo/name/unknown-service",
			"/v2/repo/name/manifests", // no required subpath
			"/v2/repo/name/manifests/not-exists",
		}
		for _, p := range paths {
			res, err := s.Client().Get(s.URL + p)
			require.NoError(err)
			require.Equal(http.StatusNotFound, res.StatusCode)
		}
	})

	t.Run("/repo/name/tags/list returns all tags", func(t *testing.T) {
		require := require.New(t)

		tags, err := repo.Tags(ctx).All(ctx)
		require.NoError(err)

		res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/tags/list", named.Name()))
		require.NoError(err)
		require.Equal(http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		require.NoError(err)

		var data struct {
			Name string
			Tags []string
		}

		err = json.Unmarshal(body, &data)
		require.NoError(err)
		require.Equal(named.Name(), data.Name)
		require.ElementsMatch(tags, data.Tags)
	})

	t.Run("/repo/name/manifests/ref", func(t *testing.T) {
		t.Run("HEAD responses OK with empty body", func(t *testing.T) {
			require := require.New(t)

			res, err := s.Client().Head(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", named.Name(), tag))
			require.NoError(err)
			require.Equal(http.StatusOK, res.StatusCode)
			require.Equal(desc.Digest.String(), res.Header.Get("Docker-Content-Digest"))

			body, err := io.ReadAll(res.Body)
			require.NoError(err)
			require.Empty(body)
		})

		t.Run("GET by digest returns manifest", func(t *testing.T) {
			require := require.New(t)

			res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", named.Name(), desc.Digest.String()))
			require.NoError(err)
			require.Equal(http.StatusOK, res.StatusCode)
			require.Equal(desc.MediaType, res.Header.Get("Content-type"))
			require.Equal(fmt.Sprint(desc.Size), res.Header.Get("Content-Length"))
		})

		t.Run("GET by tag returns manifest", func(t *testing.T) {
			require := require.New(t)

			res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", named.Name(), tag))
			require.NoError(err)
			require.Equal(http.StatusOK, res.StatusCode)
			require.Equal(desc.MediaType, res.Header.Get("Content-type"))
			require.Equal(fmt.Sprint(desc.Size), res.Header.Get("Content-Length"))
		})

		t.Run("fails if", func(t *testing.T) {
			t.Run("GET by digest with unsupported algorithm returns 501", func(t *testing.T) {
				require := require.New(t)

				res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", named.Name(), "awesome21:cool"))
				require.NoError(err)
				require.Equal(http.StatusNotImplemented, res.StatusCode)
			})

			t.Run("GET by invalid digest returns 400", func(t *testing.T) {
				require := require.New(t)

				res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", named.Name(), "sha256:vEryAweSomE"))
				require.NoError(err)
				require.Equal(http.StatusBadRequest, res.StatusCode)
			})
		})
	})
}

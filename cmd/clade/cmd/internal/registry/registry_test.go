package registry_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

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

func Test_V2(t *testing.T) {
	reg := registry.NewRegistry(t)
	reg.Repos["spider-man/tom"] = &registry.Repository{
		Name: "spider-man/tom",
		Manifests: map[string]registry.Manifest{
			"no-way-home": {
				ContentType: "movie",
				Blob:        []byte("great power"),
			},
		},
	}

	s := httptest.NewServer(reg.Handler())
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

		res, err := s.Client().Get(s.URL + "?scope=lesomnus")
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
			"/v2/resource/name/not-exists",
			"/v2/spider-man/tom/unknown-service",
			"/v2/spider-man/tom/manifests", // no required subpath
			"/v2/spider-man/tom/manifests/not-exists",
		}
		for _, p := range paths {
			res, err := s.Client().Get(s.URL + p)
			require.NoError(err)
			require.Equal(http.StatusNotFound, res.StatusCode)
		}
	})

	t.Run("/repo/name/tags/list returns all tags", func(t *testing.T) {
		require := require.New(t)

		name := "spider-man/tom"

		res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/tags/list", name))
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
		require.Contains(data.Name, name)
		require.ElementsMatch(data.Tags, reg.Repos[name].Tags())
	})

	t.Run("/repo/name/manifests/tag", func(t *testing.T) {
		t.Run("HEAD responses OK with empty body", func(t *testing.T) {
			require := require.New(t)

			name := "spider-man/tom"
			tag := "no-way-home"

			res, err := s.Client().Head(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", name, tag))
			require.NoError(err)
			require.Equal(http.StatusOK, res.StatusCode)

			body, err := io.ReadAll(res.Body)
			require.NoError(err)
			require.Empty(body)
		})

		t.Run("GET returns manifest", func(t *testing.T) {
			require := require.New(t)

			name := "spider-man/tom"
			tag := "no-way-home"

			res, err := s.Client().Get(s.URL + fmt.Sprintf("/v2/%s/manifests/%s", name, tag))
			require.NoError(err)
			require.Equal(http.StatusOK, res.StatusCode)
			require.Equal(reg.Repos[name].Manifests[tag].ContentType, res.Header.Get("Content-type"))
			require.Equal(strconv.Itoa(len(reg.Repos[name].Manifests[tag].Blob)), res.Header.Get("Content-Length"))
		})
	})
}

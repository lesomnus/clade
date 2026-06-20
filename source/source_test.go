package source_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lesomnus/clade/source"
)

func TestContainer(t *testing.T) {
	listed := ""
	deps := source.Deps{Tags: func(_ context.Context, repo string) ([]string, error) {
		listed = repo
		return []string{"1.0.0", "1.1.0"}, nil
	}}

	s, err := source.New("container", []byte("repo: up.io/base\n"), deps)
	if err != nil {
		t.Fatal(err)
	}
	vs, err := s.Versions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if listed != "up.io/base" {
		t.Errorf("listed repo = %q, want up.io/base", listed)
	}
	if len(vs) != 2 || vs[0] != "1.0.0" {
		t.Errorf("versions = %v, want [1.0.0 1.1.0]", vs)
	}
}

func TestContainerRequiresRepo(t *testing.T) {
	deps := source.Deps{Tags: func(context.Context, string) ([]string, error) { return nil, nil }}
	if _, err := source.New("container", nil, deps); err == nil {
		t.Fatal("expected error when repo is missing")
	}
}

func TestHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("1.2.3\n"))
	}))
	defer srv.Close()

	s, err := source.New("http", []byte("url: "+srv.URL+"\n"), source.Deps{})
	if err != nil {
		t.Fatal(err)
	}
	vs, err := s.Versions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 1 || vs[0] != "1.2.3" {
		t.Errorf("versions = %v, want [1.2.3]", vs)
	}
}

func TestHTTPNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	s, _ := source.New("http", []byte("url: "+srv.URL+"\n"), source.Deps{})
	if _, err := s.Versions(context.Background()); err == nil {
		t.Fatal("expected error on non-2xx response")
	}
}

func TestHTTPRequiresURL(t *testing.T) {
	if _, err := source.New("http", nil, source.Deps{}); err == nil {
		t.Fatal("expected error when url is missing")
	}
}

func TestUnknownKind(t *testing.T) {
	if _, err := source.New("nope", nil, source.Deps{}); err == nil {
		t.Fatal("expected error for unknown kind")
	}
}

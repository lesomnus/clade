package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

func init() {
	Register("http", newHTTP)
}

// httpConfig is the config for the http source.
//
//	source:
//	  kind: http
//	  url: https://downloads.claude.ai/claude-code-releases/stable
type httpConfig struct {
	URL string `yaml:"url"`
}

// httpSource fetches a single version string from an HTTP endpoint whose body is
// a bare version (e.g. "1.2.3").
type httpSource struct {
	url    string
	client *http.Client
}

func newHTTP(params []byte, _ Deps) (Source, error) {
	cfg := httpConfig{}
	if len(params) > 0 {
		if err := yaml.Unmarshal(params, &cfg); err != nil {
			return nil, fmt.Errorf("decode http source: %w", err)
		}
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("http source: url is required")
	}
	return &httpSource{url: cfg.URL, client: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (s *httpSource) Versions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", s.url, err)
	}

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", s.url, err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("get %s: unexpected status %s", s.url, res.Status)
	}

	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<16))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", s.url, err)
	}

	version := strings.TrimSpace(string(body))
	if version == "" {
		return nil, fmt.Errorf("get %s: empty body", s.url)
	}
	return []string{version}, nil
}

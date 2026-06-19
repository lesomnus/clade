package builder

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/goccy/go-yaml"
)

// options are the buildx-family build options shared by the "build" and "bake"
// strategies, decoded from a port's raw build config. Fields irrelevant to the
// decode (repo, tag, kind) are simply ignored.
type options struct {
	Dockerfile  string            `yaml:"dockerfile"`
	Context     string            `yaml:"context"`
	Target      string            `yaml:"target"`
	Platforms   []string          `yaml:"platforms"`
	Args        map[string]string `yaml:"args"`
	Labels      map[string]string `yaml:"labels"`
	Annotations []string          `yaml:"annotations"`
	CacheFrom   []string          `yaml:"cache-from"`
	CacheTo     []string          `yaml:"cache-to"`
	Secrets     []string          `yaml:"secrets"`
	SSH         []string          `yaml:"ssh"`
	NoCache     bool              `yaml:"no-cache"`
	Pull        bool              `yaml:"pull"`
	Provenance  string            `yaml:"provenance"`
	SBOM        string            `yaml:"sbom"`
	Network     string            `yaml:"network"`
	AddHosts    []string          `yaml:"add-hosts"`
	Allow       []string          `yaml:"allow"`
	ExtraArgs   []string          `yaml:"extra-args"`
}

func parseOptions(params []byte) (options, error) {
	var o options
	if len(params) > 0 {
		if err := yaml.Unmarshal(params, &o); err != nil {
			return options{}, fmt.Errorf("decode build options: %w", err)
		}
	}
	return o, nil
}

// contextDir resolves the build context against the port directory.
func (o options) contextDir(dir string) string {
	c := o.Context
	if c == "" {
		c = "."
	}
	return filepath.Join(dir, c)
}

// dockerfilePath resolves the Dockerfile against the port directory.
func (o options) dockerfilePath(dir string) string {
	f := o.Dockerfile
	if f == "" {
		f = "Dockerfile"
	}
	if filepath.IsAbs(f) {
		return f
	}
	return filepath.Join(dir, f)
}

// buildArgs merges the configured build args with the injected BASE.
func (o options) buildArgs(spec Spec) map[string]string {
	m := make(map[string]string, len(o.Args)+1)
	for k, v := range o.Args {
		m[k] = v
	}
	m["BASE"] = spec.Base
	return m
}

// imageLabels merges the configured labels with the spec's injected labels
// (the latter win on conflict).
func (o options) imageLabels(spec Spec) map[string]string {
	m := make(map[string]string, len(o.Labels)+len(spec.Labels))
	for k, v := range o.Labels {
		m[k] = v
	}
	for k, v := range spec.Labels {
		m[k] = v
	}
	return m
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

package config

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// CacheConfig configures the registry metadata cache.
type CacheConfig struct {
	// Dir is the cache directory. Empty means "<user cache dir>/clade".
	Dir string `yaml:"dir"`
	// TTL is how long cached entries are reused, as a Go duration string
	// (e.g. "24h"). Empty falls back to the default.
	TTL string `yaml:"ttl"`
}

// BuildConfig configures how images are built. The build strategy itself is
// selected per port via build.kind in port.yaml.
type BuildConfig struct {
	// Docker is the docker binary to invoke (default "docker").
	Docker string `yaml:"docker"`
}

// CompareConfig selects the outdated-comparison strategy. The strategy specific
// parameters are kept as raw YAML so the compare package can decode them.
type CompareConfig struct {
	// Kind names the comparison strategy, e.g. "created" or "digest".
	Kind string
	// Params is the raw YAML of the whole compare mapping (including kind).
	Params []byte
}

// UnmarshalYAML implements goccy/go-yaml's BytesUnmarshaler.
func (c *CompareConfig) UnmarshalYAML(b []byte) error {
	var head struct {
		Kind string `yaml:"kind"`
	}
	if err := yaml.Unmarshal(b, &head); err != nil {
		return fmt.Errorf("decode compare: %w", err)
	}

	c.Kind = head.Kind
	c.Params = b
	return nil
}

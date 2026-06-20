package config

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

// Outdated comparison is configured per port (port.yaml's compare list), not
// globally, so there is no compare config here.

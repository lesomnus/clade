package config

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/lesomnus/z"
)

var DefaultConfigPaths = []string{
	"clade.yaml",
	"clade.yml",
}

type Config struct {
	path string

	Greet GreetConfig

	Otel OtelConfig

	// Ports is the directory that holds port definitions.
	Ports string `yaml:"ports"`

	Cache CacheConfig `yaml:"cache"`

	Compare CompareConfig `yaml:"compare"`
}

func ReadFromFile(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, z.Err(err, "open")
	}

	var c Config
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return nil, z.Err(err, "decode")
	}

	c.path = p
	return &c, nil
}

func (c *Config) Path() string {
	return c.path
}

func (c *Config) Evaluate() error {
	z.FallbackP(&c.Greet.Format, "Hello, %s!")
	z.FallbackP(&c.Ports, "ports")
	z.FallbackP(&c.Cache.TTL, "24h")
	z.FallbackP(&c.Compare.Kind, "created")
	return nil
}

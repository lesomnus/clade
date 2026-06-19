package compare

import (
	"context"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/lesomnus/clade/registry"
)

// DefaultBaseDigestLabel is the label that records the digest of the base image
// a target was built from. The build step is responsible for writing it.
const DefaultBaseDigestLabel = "org.opencontainers.image.base.digest"

func init() {
	Register("digest", newDigest)
}

// digestConfig is the optional config for the digest strategy.
//
//	compare:
//	  kind: digest
//	  label: org.opencontainers.image.base.digest
type digestConfig struct {
	Label string `yaml:"label"`
}

// digest compares the base digest recorded on the target (as a label) with the
// current digest of the base image. They differ when the base has moved on.
type digest struct {
	label string
}

func newDigest(params []byte) (Comparator, error) {
	cfg := digestConfig{}
	if len(params) > 0 {
		if err := yaml.Unmarshal(params, &cfg); err != nil {
			return nil, fmt.Errorf("decode digest config: %w", err)
		}
	}
	if cfg.Label == "" {
		cfg.Label = DefaultBaseDigestLabel
	}
	return digest{label: cfg.Label}, nil
}

func (d digest) IsOutdated(_ context.Context, base, target *registry.ImageInfo) (bool, error) {
	return target.Labels[d.label] != base.Digest, nil
}

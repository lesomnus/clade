package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func init() {
	Register("bake", newBake)
}

// bake builds via `docker buildx bake`, synthesizing a single-target bake
// definition from the options and spec.
type bake struct {
	opts options
	spec Spec
}

func newBake(params []byte, spec Spec) (Builder, error) {
	o, err := parseOptions(params)
	if err != nil {
		return nil, err
	}
	return &bake{opts: o, spec: spec}, nil
}

// bakeTarget mirrors the attributes of a `docker buildx bake` target.
type bakeTarget struct {
	Context     string            `json:"context,omitempty"`
	Dockerfile  string            `json:"dockerfile,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Target      string            `json:"target,omitempty"`
	Platforms   []string          `json:"platforms,omitempty"`
	Args        map[string]string `json:"args,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations []string          `json:"annotations,omitempty"`
	CacheFrom   []string          `json:"cache-from,omitempty"`
	CacheTo     []string          `json:"cache-to,omitempty"`
	Secret      []string          `json:"secret,omitempty"`
	SSH         []string          `json:"ssh,omitempty"`
	NoCache     *bool             `json:"no-cache,omitempty"`
	Pull        *bool             `json:"pull,omitempty"`
	Network     string            `json:"network,omitempty"`
	Output      []string          `json:"output,omitempty"`
}

type bakeFile struct {
	Target map[string]bakeTarget `json:"target"`
}

const bakeTargetName = "default"

func (b *bake) file() bakeFile {
	o, spec := b.opts, b.spec
	t := bakeTarget{
		Context:     o.contextDir(spec.Dir),
		Dockerfile:  o.dockerfilePath(spec.Dir),
		Tags:        spec.Tags,
		Target:      o.Target,
		Platforms:   o.Platforms,
		Args:        o.buildArgs(spec),
		Labels:      o.imageLabels(spec),
		Annotations: o.Annotations,
		CacheFrom:   o.CacheFrom,
		CacheTo:     o.CacheTo,
		Secret:      o.Secrets,
		SSH:         o.SSH,
		Network:     o.Network,
	}
	if o.NoCache {
		t.NoCache = boolPtr(true)
	}
	if o.Pull {
		t.Pull = boolPtr(true)
	}
	switch {
	case spec.Push:
		t.Output = []string{"type=registry"}
	case spec.Load:
		t.Output = []string{"type=docker"}
	}
	return bakeFile{Target: map[string]bakeTarget{bakeTargetName: t}}
}

func (b *bake) args(file string) []string {
	args := []string{"buildx", "bake", "--file", file}
	if b.opts.Provenance != "" {
		args = append(args, "--provenance", b.opts.Provenance)
	}
	if b.opts.SBOM != "" {
		args = append(args, "--sbom", b.opts.SBOM)
	}
	args = append(args, b.opts.ExtraArgs...)
	return append(args, bakeTargetName)
}

// Build implements Builder.
func (b *bake) Build(ctx context.Context) error {
	if b.spec.Push && b.spec.Load {
		return fmt.Errorf("push and load are mutually exclusive")
	}

	def, err := json.MarshalIndent(b.file(), "", "  ")
	if err != nil {
		return fmt.Errorf("encode bake definition: %w", err)
	}

	o := b.spec.execOpts()
	if o.dryRun {
		fmt.Fprintf(o.stdout, "%s\n", def)
		fmt.Fprintln(o.stdout, shellCommand(o.bin, b.args("clade-bake.json")))
		return nil
	}

	f, err := os.CreateTemp("", "clade-bake-*.json")
	if err != nil {
		return fmt.Errorf("create bake file: %w", err)
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(def); err != nil {
		f.Close()
		return fmt.Errorf("write bake file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close bake file: %w", err)
	}

	return o.run(ctx, b.args(f.Name()))
}

func boolPtr(v bool) *bool { return &v }

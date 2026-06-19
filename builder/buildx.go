package builder

import (
	"context"
	"fmt"
)

func init() {
	Register("build", newBuildx)
}

// buildx builds via `docker buildx build`.
type buildx struct {
	opts options
	spec Spec
}

func newBuildx(params []byte, spec Spec) (Builder, error) {
	o, err := parseOptions(params)
	if err != nil {
		return nil, err
	}
	return &buildx{opts: o, spec: spec}, nil
}

// Build implements Builder.
func (b *buildx) Build(ctx context.Context) error {
	if b.spec.Push && b.spec.Load {
		return fmt.Errorf("push and load are mutually exclusive")
	}
	return b.spec.execOpts().run(ctx, b.argv())
}

func (b *buildx) argv() []string {
	o, spec := b.opts, b.spec

	args := []string{"buildx", "build", "--file", o.dockerfilePath(spec.Dir)}
	if o.Target != "" {
		args = append(args, "--target", o.Target)
	}
	for _, p := range o.Platforms {
		args = append(args, "--platform", p)
	}

	build_args := o.buildArgs(spec)
	for _, k := range sortedKeys(build_args) {
		args = append(args, "--build-arg", k+"="+build_args[k])
	}
	labels := o.imageLabels(spec)
	for _, k := range sortedKeys(labels) {
		args = append(args, "--label", k+"="+labels[k])
	}

	for _, a := range o.Annotations {
		args = append(args, "--annotation", a)
	}
	for _, c := range o.CacheFrom {
		args = append(args, "--cache-from", c)
	}
	for _, c := range o.CacheTo {
		args = append(args, "--cache-to", c)
	}
	for _, s := range o.Secrets {
		args = append(args, "--secret", s)
	}
	for _, s := range o.SSH {
		args = append(args, "--ssh", s)
	}
	if o.NoCache {
		args = append(args, "--no-cache")
	}
	if o.Pull {
		args = append(args, "--pull")
	}
	if o.Provenance != "" {
		args = append(args, "--provenance", o.Provenance)
	}
	if o.SBOM != "" {
		args = append(args, "--sbom", o.SBOM)
	}
	if o.Network != "" {
		args = append(args, "--network", o.Network)
	}
	for _, h := range o.AddHosts {
		args = append(args, "--add-host", h)
	}
	for _, a := range o.Allow {
		args = append(args, "--allow", a)
	}
	for _, t := range spec.Tags {
		args = append(args, "--tag", t)
	}
	if spec.Push {
		args = append(args, "--push")
	}
	if spec.Load {
		args = append(args, "--load")
	}
	args = append(args, o.ExtraArgs...)

	return append(args, o.contextDir(spec.Dir))
}

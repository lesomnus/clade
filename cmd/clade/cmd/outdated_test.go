package cmd_test

import (
	"bytes"
	"encoding/hex"
	"os"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/lesomnus/clade/internal/registry"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

func TestOutdatedCmd(t *testing.T) {
	tcs := []struct {
		desc    string
		prepare func(t *testing.T) (*TmpPortDir, *registry.Registry)
		include []string
		exclude []string
	}{
		{
			desc: "not outdated if image has same layers with its base",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				foo, err := reference.ParseNamed("cr.io/repo/foo")
				require.NoError(err)

				foo_repo := reg.NewRepository(foo)
				foo_desc, foo_manif := foo_repo.PopulateManifest()
				foo_repo.Storage.Tags["1"] = foo_desc

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_desc, bar_manif := bar_repo.PopulateManifest()
				bar_repo.Storage.Tags["1"] = bar_desc

				foo_manif.Layers = append([]distribution.Descriptor{}, bar_manif.Layers...)
				foo_manif.Layers = append(foo_manif.Layers, distribution.Descriptor{Digest: digest.FromBytes([]byte("baz"))})

				return ports, reg
			},
			exclude: []string{
				"cr.io/repo/foo:1",
				"cr.io/repo/bar:1",
			},
		},
		{
			desc: "outdated if image has different layers with its base",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				foo, err := reference.ParseNamed("cr.io/repo/foo")
				require.NoError(err)

				foo_repo := reg.NewRepository(foo)
				foo_repo.Storage.Tags["1"], _ = foo_repo.PopulateManifest()

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_repo.Storage.Tags["1"], _ = bar_repo.PopulateManifest()

				return ports, reg
			},
			include: []string{"cr.io/repo/foo:1"},
			exclude: []string{"cr.io/repo/bar:1"},
		},
		{
			desc: "not outdated if deref ID not changed",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				foo, err := reference.ParseNamed("cr.io/repo/foo")
				require.NoError(err)

				foo_repo := reg.NewRepository(foo)
				foo_desc, foo_manif := foo_repo.PopulateOciManifest()
				foo_repo.Storage.Tags["1"] = foo_desc

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_desc, _ := bar_repo.PopulateManifest()
				bar_repo.Storage.Tags["1"] = bar_desc

				dgst, err := hex.DecodeString(bar_desc.Digest.Encoded())
				require.NoError(err)

				foo_manif.Annotations[clade.AnnotationDerefId] = clade.CalcDerefId(dgst)

				return ports, reg
			},
			exclude: []string{
				"cr.io/repo/foo:1",
				"cr.io/repo/bar:1",
			},
		},
		{
			desc: "outdated is checked by layers if deref ID is not annotated in OCI image",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				foo, err := reference.ParseNamed("cr.io/repo/foo")
				require.NoError(err)

				foo_repo := reg.NewRepository(foo)
				foo_repo.Storage.Tags["1"], _ = foo_repo.PopulateOciManifest()

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_repo.Storage.Tags["1"], _ = bar_repo.PopulateOciManifest()

				return ports, reg
			},
			include: []string{"cr.io/repo/foo:1"},
			exclude: []string{"cr.io/repo/bar:1"},
		},
		{
			desc: "outdated if the image is not exists",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_repo.Storage.Tags["1"], _ = bar_repo.PopulateOciManifest()

				return ports, reg
			},
			include: []string{"cr.io/repo/foo:1"},
			exclude: []string{"cr.io/repo/bar:1"},
		},
		{
			desc: "outdated if the tag is not exists",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				foo, err := reference.ParseNamed("cr.io/repo/foo")
				require.NoError(err)

				reg.NewRepository(foo)

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_repo.Storage.Tags["1"], _ = bar_repo.PopulateOciManifest()

				return ports, reg
			},
			include: []string{"cr.io/repo/foo:1"},
			exclude: []string{"cr.io/repo/bar:1"},
		},
		{
			desc: "only first tag is printed",
			prepare: func(t *testing.T) (*TmpPortDir, *registry.Registry) {
				require := require.New(t)

				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1, 2]
    from: cr.io/repo/bar:1`)

				reg := registry.NewRegistry()

				foo, err := reference.ParseNamed("cr.io/repo/foo")
				require.NoError(err)

				reg.NewRepository(foo)

				bar, err := reference.ParseNamed("cr.io/repo/bar")
				require.NoError(err)

				bar_repo := reg.NewRepository(bar)
				bar_repo.Storage.Tags["1"], _ = bar_repo.PopulateOciManifest()

				return ports, reg
			},
			include: []string{"cr.io/repo/foo:1"},
			exclude: []string{
				"cr.io/repo/foo:2",
				"cr.io/repo/bar:1",
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			ports, reg := tc.prepare(t)

			buff := new(bytes.Buffer)
			svc := cmd.NewCmdService()
			svc.Out = buff
			svc.RegistryClient = reg

			flags := cmd.OutdatedFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: ports.Dir,
				},
			}

			c := cmd.CreateOutdatedCmd(&flags, svc)
			c.SetOut(os.Stderr)

			err := c.Execute()
			require.NoError(err)

			output := buff.String()
			for _, s := range tc.include {
				require.Contains(output, s)
			}
			for _, s := range tc.exclude {
				require.NotContains(output, s)
			}
		})
	}
}

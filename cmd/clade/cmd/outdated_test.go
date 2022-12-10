package cmd_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/distribution/distribution/v3"
	"github.com/distribution/distribution/v3/reference"
	"github.com/distribution/distribution/v3/registry/api/errcode"
	v2 "github.com/distribution/distribution/v3/registry/api/v2"
	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
)

type LayerService struct {
	cmd.Service
	Layers map[string][]distribution.Descriptor
}

func (s *LayerService) GetLayer(ctx context.Context, named_tagged reference.NamedTagged) ([]distribution.Descriptor, error) {
	layers, ok := s.Layers[named_tagged.String()]
	if !ok {
		errs := make(errcode.Errors, 1)
		errs[0] = v2.ErrorCodeManifestUnknown
		return nil, errs
	}

	return layers, nil
}

func TestOutdatedCmd(t *testing.T) {
	ports := GenerateSamplePorts(t)

	newLayers := func() map[string][]distribution.Descriptor {
		return map[string][]distribution.Descriptor{
			"registry.hub.docker.com/library/gcc:12.2": {
				{Digest: digest.Digest("a")},
			},
			"ghcr.io/lesomnus/gcc:12.2": {
				{Digest: digest.Digest("a")},
				{Digest: digest.Digest("b")},
			},
			"ghcr.io/lesomnus/ffmpeg:4.4.1": {
				{Digest: digest.Digest("a")},
				{Digest: digest.Digest("b")},
				{Digest: digest.Digest("f")},
			},
			"ghcr.io/lesomnus/pcl:1.11.1": {
				{Digest: digest.Digest("a")},
				{Digest: digest.Digest("b")},
				{Digest: digest.Digest("c")},
			},
			"registry.hub.docker.com/library/node:19": {
				{Digest: digest.Digest("a")},
			},
			"ghcr.io/lesomnus/node:19": {
				{Digest: digest.Digest("a")},
				{Digest: digest.Digest("b")},
			},
		}
	}

	tcs := []struct {
		desc    string
		args    []string
		layers  map[string][]distribution.Descriptor
		include []string
		exclude []string
	}{
		{
			desc:   "prints nothing if all up-to-date",
			layers: newLayers(),
			exclude: []string{
				"registry.hub.docker.com/library/gcc:12.2",
				"ghcr.io/lesomnus/gcc:12.2",
				"ghcr.io/lesomnus/pcl:1.11.1",
				"registry.hub.docker.com/library/node:19",
				"ghcr.io/lesomnus/node:19",
			},
		},
		{
			desc: "outdated if child does not have parent as base",
			layers: (func() map[string][]distribution.Descriptor {
				layers := newLayers()
				layers["registry.hub.docker.com/library/node:19"][0].Digest = "a2"
				return layers
			})(),
			include: []string{
				"ghcr.io/lesomnus/node:19",
			},
			exclude: []string{
				"registry.hub.docker.com/library/gcc:12.2",
				"ghcr.io/lesomnus/gcc:12.2",
				"ghcr.io/lesomnus/pcl:1.11.1",
				"registry.hub.docker.com/library/node:19",
			},
		},
		{
			desc: "child image of outdated image is not printed",
			layers: (func() map[string][]distribution.Descriptor {
				layers := newLayers()
				layers["registry.hub.docker.com/library/gcc:12.2"][0].Digest = "a2"
				return layers
			})(),
			include: []string{
				"ghcr.io/lesomnus/gcc:12.2",
			},
			exclude: []string{
				"registry.hub.docker.com/library/gcc:12.2",
				"ghcr.io/lesomnus/pcl:1.11.1",
				"registry.hub.docker.com/library/node:19",
				"ghcr.io/lesomnus/node:19",
			},
		},
		{
			desc: "outdated if layer manifest does not eixsts",
			layers: (func() map[string][]distribution.Descriptor {
				layers := newLayers()
				delete(layers, "ghcr.io/lesomnus/pcl:1.11.1")
				return layers
			})(),
			include: []string{
				"ghcr.io/lesomnus/pcl:1.11.1",
			},
			exclude: []string{
				"registry.hub.docker.com/library/gcc:12.2",
				"ghcr.io/lesomnus/gcc:12.2",
				"registry.hub.docker.com/library/node:19",
				"ghcr.io/lesomnus/node:19",
			},
		},
		{
			desc:   "skipped image is not printed",
			layers: newLayers(),
			exclude: []string{
				`ghcr.io/lesomnus/skipped:42`,
				`ghcr.io/lesomnus/skipped-child:36`,
			},
		},
		{
			desc: "--all flag prints all images including skipped images",
			layers: (func() map[string][]distribution.Descriptor {
				layers := newLayers()
				layers["ghcr.io/lesomnus/skipped:42"] = []distribution.Descriptor{
					{Digest: digest.Digest("a")},
					{Digest: digest.Digest("b")},
					{Digest: digest.Digest("s")},
				}

				return layers
			})(),
			args: []string{"--all"},
			include: []string{
				`ghcr.io/lesomnus/skipped-child:36`,
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)
			buff := new(bytes.Buffer)

			svc := cmd.NewCmdService()
			svc.Sink = buff
			layer_svc := &LayerService{
				Service: svc,
				Layers:  tc.layers,
			}
			flags := cmd.OutdatedFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: ports,
				},
			}

			c := cmd.CreateOutdatedCmd(&flags, layer_svc)
			c.SetOut(io.Discard)
			if tc.args != nil {
				c.SetArgs(tc.args)
			}

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

	t.Run("fails if", func(t *testing.T) {
		t.Run("ports directory does not exist", func(t *testing.T) {
			require := require.New(t)

			svc := cmd.NewCmdService()
			flags := cmd.OutdatedFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: "not-exists",
				},
			}

			c := cmd.CreateOutdatedCmd(&flags, svc)
			c.SetOutput(io.Discard)
			err := c.Execute()
			require.ErrorContains(err, "no such file or directory")
		})

		t.Run("manifests of top-level image does not exist", func(t *testing.T) {
			require := require.New(t)

			svc := cmd.NewCmdService()
			svc.Sink = io.Discard
			layer_svc := &LayerService{
				Service: svc,
				Layers:  newLayers(),
			}
			flags := cmd.OutdatedFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: ports,
				},
			}

			delete(layer_svc.Layers, "registry.hub.docker.com/library/gcc:12.2")

			c := cmd.CreateOutdatedCmd(&flags, layer_svc)
			c.SetOutput(io.Discard)
			err := c.Execute()
			require.ErrorContains(err, "failed to get layers of")
			require.ErrorContains(err, "registry.hub.docker.com/library/gcc:12.2")
		})
	})
}

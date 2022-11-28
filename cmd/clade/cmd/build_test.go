package cmd_test

import (
	"io"
	"testing"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/builder"
	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/stretchr/testify/require"
)

type MockBuilder struct {
	Image *clade.ResolvedImage
}

func (b *MockBuilder) Build(image *clade.ResolvedImage) error {
	b.Image = image
	return nil
}

func TestBuildCmd(t *testing.T) {
	ports := GenerateSamplePorts(t)

	t.Run("registered builder is invoked with arguments", func(t *testing.T) {
		require := require.New(t)

		svc := cmd.NewCmdService()
		svc.Sink = io.Discard
		flags := cmd.BuildFlags{
			RootFlags: &cmd.RootFlags{
				PortsPath: ports,
			},
		}

		var build_config builder.BuilderConfig

		name := t.TempDir()
		b := MockBuilder{}
		builder.Register(name, func(conf builder.BuilderConfig) (builder.Builder, error) {
			build_config = conf
			return &b, nil
		})

		c := cmd.CreateBuildCmd(&flags, svc)
		c.SetArgs([]string{"--builder", name, "--dry-run", "ghcr.io/lesomnus/pcl:1.11", "--", "--some-arg"})
		err := c.Execute()
		require.NoError(err)
		require.True(build_config.DryRun)
		require.Equal([]string{"--some-arg"}, build_config.Args)
		require.Equal("ghcr.io/lesomnus/pcl", b.Image.Named.String())
	})

	t.Run("fails if", func(t *testing.T) {
		builder_name := t.TempDir()
		b := MockBuilder{}
		builder.Register(builder_name, func(conf builder.BuilderConfig) (builder.Builder, error) {
			return &b, nil
		})

		tcs := []struct {
			desc string
			args []string
			msgs []string
		}{
			{
				desc: "no argument given",
				args: []string{},
				msgs: []string{"requires", "1 arg"},
			},
			{
				desc: "argument is invalid reference",
				args: []string{"ghcr/lesomnus/pcl"},
				msgs: []string{"must be canonical"},
			},
			{
				desc: "argument is un-tagged reference",
				args: []string{"ghcr.io/lesomnus/pcl"},
				msgs: []string{"must be tagged"},
			},
			{
				desc: "builder does not exist",
				args: []string{"--builder", "not-exists", "ghcr.io/lesomnus/java:19"},
				msgs: []string{"builder", "not exists"},
			},
			{
				desc: "reference does not exist in the ports",
				args: []string{"--builder", builder_name, "ghcr.io/lesomnus/java:19"},
				msgs: []string{"failed to find"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				svc := cmd.NewCmdService()
				svc.Sink = io.Discard
				flags := cmd.BuildFlags{
					RootFlags: &cmd.RootFlags{
						PortsPath: ports,
					},
				}

				c := cmd.CreateBuildCmd(&flags, svc)
				c.SetOutput(io.Discard)
				c.SetArgs(tc.args)
				err := c.Execute()
				require.Error(err)

				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}

		t.Run("ports directory does not exist", func(t *testing.T) {
			require := require.New(t)

			svc := cmd.NewCmdService()
			svc.Sink = io.Discard
			flags := cmd.BuildFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: "not-exists",
				},
			}

			c := cmd.CreateBuildCmd(&flags, svc)
			c.SetOutput(io.Discard)
			c.SetArgs([]string{"--builder", builder_name, "ghcr.io/lesomnus/java:19"})
			err := c.Execute()
			require.ErrorContains(err, "no such file or directory")
		})
	})
}

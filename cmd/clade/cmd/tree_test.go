package cmd_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/stretchr/testify/require"
)

// type MockService struct {
// 	Sink io.Writer
// }

func TestTreeCmd(t *testing.T) {
	ports := GenerateSamplePorts(t)

	t.Run("list ports in tree view", func(t *testing.T) {
		tcs := []struct {
			desc    string
			args    []string
			include []string
			exclude []string
		}{
			{
				desc: "different tags of the same image are output together with their full names",
				args: []string{},
				include: []string{
					`registry.hub.docker.com/library/gcc:12.2
	ghcr.io/lesomnus/gcc:12.2
	ghcr.io/lesomnus/gcc:12
		ghcr.io/lesomnus/pcl:1.11.1
		ghcr.io/lesomnus/pcl:1.11`,
				},
			},
			{
				desc: "--strip flag omits first N levels",
				args: []string{"--strip", "1"},
				include: []string{
					`ghcr.io/lesomnus/gcc:12.2
ghcr.io/lesomnus/gcc:12
	ghcr.io/lesomnus/pcl:1.11.1
	ghcr.io/lesomnus/pcl:1.11`,
				},
			},
			{
				desc: "--depth flag prints first N levels only",
				args: []string{"--depth", "1"},
				include: []string{
					"registry.hub.docker.com/library/gcc:12.2",
				},
				exclude: []string{
					"ghcr.io/lesomnus/gcc:12.2",
				},
			},
			{
				desc: "--fold flag prints only major tag",
				args: []string{"--fold"},
				include: []string{
					`registry.hub.docker.com/library/gcc:12.2
	ghcr.io/lesomnus/gcc:12.2
		ghcr.io/lesomnus/pcl:1.11.1`,
				},
			},
			{
				desc: "print sub-tree",
				args: []string{"ghcr.io/lesomnus/gcc:12.2"},
				include: []string{
					`ghcr.io/lesomnus/gcc:12.2
ghcr.io/lesomnus/gcc:12
	ghcr.io/lesomnus/pcl:1.11.1
	ghcr.io/lesomnus/pcl:1.11`,
				},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)
				buff := new(bytes.Buffer)

				svc := cmd.NewCmdService()
				svc.Sink = buff
				flags := cmd.TreeFlags{
					RootFlags: &cmd.RootFlags{
						PortsPath: ports,
					},
				}

				c := cmd.CreateTreeCmd(&flags, svc)
				c.SetArgs(tc.args)
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
	})

	t.Run("fails if", func(t *testing.T) {
		t.Run("ports directory does not exist", func(t *testing.T) {
			require := require.New(t)

			svc := cmd.NewCmdService()
			flags := cmd.TreeFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: "not-exists",
				},
			}

			c := cmd.CreateTreeCmd(&flags, svc)
			c.SetOutput(io.Discard)
			err := c.Execute()
			require.ErrorContains(err, "no such file or directory")
		})

		t.Run("sub-tree does not exist", func(t *testing.T) {
			require := require.New(t)

			svc := cmd.NewCmdService()
			flags := cmd.TreeFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: ports,
				},
			}

			c := cmd.CreateTreeCmd(&flags, svc)
			c.SetOutput(io.Discard)
			c.SetArgs([]string{"cr.io/somewhere/not-exists:tag"})
			err := c.Execute()
			require.ErrorContains(err, "not-exists:tag not found")
		})
	})
}

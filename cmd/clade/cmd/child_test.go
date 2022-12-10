package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/stretchr/testify/require"
)

func TestChildCmd(t *testing.T) {
	ports := GenerateSamplePorts(t)

	t.Run("list child references", func(t *testing.T) {
		require := require.New(t)
		buff := new(bytes.Buffer)

		svc := cmd.NewCmdService()
		svc.Sink = buff
		flags := cmd.TreeFlags{
			RootFlags: &cmd.RootFlags{
				PortsPath: ports,
			},
		}

		c := cmd.CreateChildCmd(&flags, cmd.CreateTreeCmd(&flags, svc))

		c.SetArgs([]string{"ghcr.io/lesomnus/gcc:12"})
		err := c.Execute()
		require.NoError(err)

		output := buff.String()
		refs := strings.Split(output, "\n")
		if refs[len(refs)-1] == "" {
			refs = refs[0 : len(refs)-1]
		}

		require.ElementsMatch(refs, []string{
			"ghcr.io/lesomnus/ffmpeg:4.4.1",
			"ghcr.io/lesomnus/pcl:1.11.1",
		})
	})
}

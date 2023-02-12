package cmd_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/stretchr/testify/require"
)

func TestCachePathCmd(t *testing.T) {
	require := require.New(t)
	buff := new(bytes.Buffer)

	svc := cmd.NewCmdService()
	svc.Sink = buff

	c := cmd.CreateCachePathCmd(svc)
	c.SetOutput(io.Discard)
	err := c.Execute()
	require.NoError(err)
	require.Equal(cmd.RegistryCache.Root, strings.TrimSpace(buff.String()))
}

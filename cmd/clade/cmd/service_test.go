package cmd_test

import (
	"os"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	t.Run("default output is standard output", func(t *testing.T) {
		require := require.New(t)

		svc := cmd.NewCmdService()
		require.Equal(os.Stdout, svc.Output())
	})
}

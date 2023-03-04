package cmd_test

import (
	"context"
	"os"
	"testing"

	"github.com/lesomnus/clade"
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

func TestServiceLoadBuildTreeFromFs(t *testing.T) {
	svc := cmd.NewCmdService()

	t.Run("load build tree from the ports directory", func(t *testing.T) {
		require := require.New(t)

		ports := GenerateSamplePorts(t)

		ctx := context.Background()
		bt := clade.NewBuildTree()
		err := svc.LoadBuildTreeFromFs(ctx, bt, ports)
		require.NoError(err)
		require.Greater(len(bt.Tree), 0) // parent and child
	})

	t.Run("fails if directory does not exists", func(t *testing.T) {
		require := require.New(t)

		ctx := context.Background()
		bt := clade.NewBuildTree()
		err := svc.LoadBuildTreeFromFs(ctx, bt, "not-exists")
		require.ErrorContains(err, "not-exists")
		require.ErrorContains(err, "no such file or directory")
	})
}

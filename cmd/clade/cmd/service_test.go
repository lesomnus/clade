package cmd_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
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

func TestServiceGetLayers(t *testing.T) {
	reg := registry.NewRegistry(t)

	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_rul, err := url.Parse(s.URL)
	require.NoError(t, err)

	named, err := reference.ParseNamed(reg_rul.Host + "/repo/name")
	require.NoError(t, err)

	name := reference.Path(named)

	reg.Repos[name] = &registry.Repository{
		Name:      name,
		Manifests: registry.SampleManifests,
	}

	reg_client := client.NewDistRegistry()
	reg_client.Transport = s.Client().Transport

	svc := cmd.NewCmdService()
	svc.Loader.Expander.Registry = reg_client

	t.Run("gets layers of the given tag", func(t *testing.T) {
		require := require.New(t)

		tagged, err := reference.WithTag(named, "foo")
		require.NoError(err)

		ctx := context.Background()
		layers, err := svc.GetLayer(ctx, tagged)
		require.NoError(err)
		require.Len(layers, 3)
	})
}

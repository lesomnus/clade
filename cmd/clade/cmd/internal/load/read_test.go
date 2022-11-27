package load_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
	"github.com/stretchr/testify/require"
)

func TestReadPorts(t *testing.T) {
	ctx := context.Background()

	t.Run("reads port.yaml for each directory", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()

		// Make valid port directory.
		err := os.Mkdir(filepath.Join(tmp, "foo"), 0755)
		require.NoError(err)

		err = os.WriteFile(
			filepath.Join(tmp, "foo", "port.yaml"),
			[]byte(`name: ghcr.io/repo/name`),
			0644)
		require.NoError(err)

		// Directory without port.yaml.
		err = os.Mkdir(filepath.Join(tmp, "bar"), 0755)
		require.NoError(err)

		err = os.WriteFile(filepath.Join(tmp, "bar", "secrets"), []byte("Frank Moses"), 0644)
		require.NoError(err)

		// A file.
		err = os.WriteFile(filepath.Join(tmp, "baz"), []byte("Sarah Ross"), 0644)
		require.NoError(err)

		ports, err := load.ReadFromFs(ctx, tmp)
		require.NoError(err)
		require.Len(ports, 1)
	})

	t.Run("fails if directory not exists", func(t *testing.T) {
		require := require.New(t)

		_, err := load.ReadFromFs(ctx, "/not exists")
		require.ErrorIs(err, os.ErrNotExist)
	})

	t.Run("fails if port.yaml is invalid", func(t *testing.T) {
		require := require.New(t)

		tmp := t.TempDir()

		// Make valid port directory.
		err := os.Mkdir(filepath.Join(tmp, "foo"), 0755)
		require.NoError(err)

		err = os.WriteFile(
			filepath.Join(tmp, "foo", "port.yaml"),
			[]byte(`name: invalid name`),
			0644)
		require.NoError(err)

		_, err = load.ReadFromFs(ctx, tmp)
		require.ErrorContains(err, filepath.Join(tmp, "foo", "port.yaml"))
		require.ErrorContains(err, "name")
		require.ErrorContains(err, "invalid")
	})
}

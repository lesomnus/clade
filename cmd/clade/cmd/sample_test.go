package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func GenerateSamplePorts(t *testing.T) string {
	require := require.New(t)

	tmp := t.TempDir()

	err := os.Mkdir(filepath.Join(tmp, "node"), 0755)
	require.NoError(err)

	err = os.WriteFile(
		filepath.Join(tmp, "node", "port.yaml"),
		[]byte(`
name: ghcr.io/lesomnus/node
images:
  - tags: [19]
    from: registry.hub.docker.com/library/node:19
`), 0644)
	require.NoError(err)

	err = os.Mkdir(filepath.Join(tmp, "gcc"), 0755)
	require.NoError(err)

	err = os.WriteFile(
		filepath.Join(tmp, "gcc", "port.yaml"),
		[]byte(`
name: ghcr.io/lesomnus/gcc
images:
  - tags: [12.2, 12]
    from: registry.hub.docker.com/library/gcc:12.2
`), 0644)
	require.NoError(err)

	err = os.Mkdir(filepath.Join(tmp, "pcl"), 0755)
	require.NoError(err)

	err = os.WriteFile(
		filepath.Join(tmp, "pcl", "port.yaml"),
		[]byte(`
name: ghcr.io/lesomnus/pcl
images:
  - tags: [1.11.1, 1.11]
    from: ghcr.io/lesomnus/gcc:12.2
`), 0644)
	require.NoError(err)

	return tmp
}

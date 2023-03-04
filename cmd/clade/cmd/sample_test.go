package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type TmpPortDir struct {
	T   *testing.T
	Dir string
}

func NewTmpPortDir(t *testing.T) *TmpPortDir {
	return &TmpPortDir{
		T:   t,
		Dir: t.TempDir(),
	}
}

func (d *TmpPortDir) AddRaw(name string, port string) {
	require := require.New(d.T)

	err := os.Mkdir(filepath.Join(d.Dir, name), 0755)
	require.NoError(err)

	err = os.WriteFile(filepath.Join(d.Dir, name, "port.yaml"), []byte(port), 0644)
	require.NoError(err)
}

func GenerateSamplePorts(t *testing.T) string {
	require := require.New(t)

	tmp := t.TempDir()
	add := func(name string, port string) {
		err := os.Mkdir(filepath.Join(tmp, name), 0755)
		require.NoError(err)

		err = os.WriteFile(filepath.Join(tmp, name, "port.yaml"), []byte(port), 0644)
		require.NoError(err)
	}

	add("node", `
name: ghcr.io/lesomnus/node
images:
  - tags: [19]
    from: registry.hub.docker.com/library/node:19
`)

	add("gcc", `
name: ghcr.io/lesomnus/gcc
images:
  - tags: [12.2, 12]
    from: registry.hub.docker.com/library/gcc:12.2
`)

	add("pcl", `
name: ghcr.io/lesomnus/pcl
images:
  - tags: [1.11.1, 1.11]
    from: ghcr.io/lesomnus/gcc:12.2
`)

	add("ffmpeg", `
name: ghcr.io/lesomnus/ffmpeg
images:
  - tags: [4.4.1, 4.4]
    from: ghcr.io/lesomnus/gcc:12.2
`)

	add("skipped", `
name: ghcr.io/lesomnus/skipped
skip: true
images:
  - tags: [42]
    from: ghcr.io/lesomnus/gcc:12.2
`)

	add("skipped-child", `
name: ghcr.io/lesomnus/skipped-child
images:
  - tags: [36]
    from: ghcr.io/lesomnus/skipped:42
`)

	return tmp
}

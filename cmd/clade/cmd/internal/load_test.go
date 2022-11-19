package internal_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/lesomnus/clade/tree"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestExpandImage(t *testing.T) {
	empty_manifest := manifest{
		contentType: "application/vnd.docker.distribution.manifest.list.v2+json",
		blob:        []byte{},
	}

	reg := newRegistry(t)
	reg.repos = map[string]*repository{
		"repo/single": {
			name: "repo/single",
			manifests: map[string]manifest{
				"1.0.0": empty_manifest,
			},
		},
		"repo/patched": {
			name: "repo/patched",
			manifests: map[string]manifest{
				"1.0.0": empty_manifest,
				"1.0.1": empty_manifest,
			},
		},
	}

	s := httptest.NewTLSServer(reg.handler())
	defer s.Close()

	u, err := url.Parse(s.URL)
	require.NoError(t, err)

	image_name, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	t.Run("from static tag", func(t *testing.T) {
		require := require.New(t)

		named, err := reference.ParseNamed(fmt.Sprintf("%s/repo/single", u.Host))
		require.NoError(err)

		image := &clade.Image{Named: image_name}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [foo]
from:
  name: %s
  tag: "1.0.0"`, named.String())), image)
		require.NoError(err)

		err = internal.Cache.Clear()
		require.NoError(err)

		ctx := context.Background()
		images, err := internal.ExpandImage(ctx, image, clade.NewBuildTree(), internal.WithTransport(s.Client().Transport))
		require.NoError(err)
		require.Len(images[0].Tags, 1)
		require.Equal("foo", images[0].Tags[0])
		require.Equal("1.0.0", images[0].From.Tag())
	})

	t.Run("executes pipeline", func(t *testing.T) {
		require := require.New(t)

		named, err := reference.ParseNamed(fmt.Sprintf("%s/repo/single", u.Host))
		require.NoError(err)

		image := &clade.Image{Named: image_name}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [foo]
from:
  name: %s
  tag: ( tags | semver )`, named.String())), image)
		require.NoError(err)

		err = internal.Cache.Clear()
		require.NoError(err)

		ctx := context.Background()
		images, err := internal.ExpandImage(ctx, image, clade.NewBuildTree(), internal.WithTransport(s.Client().Transport))
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("foo", images[0].Tags[0])
		require.Equal("1.0.0", images[0].From.Tag())
	})

	t.Run("expands images as many as remote tags", func(t *testing.T) {
		require := require.New(t)

		named, err := reference.ParseNamed(fmt.Sprintf("%s/repo/patched", u.Host))
		require.NoError(err)

		image := &clade.Image{Named: image_name}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s
  tag: ( tags | semver )`, named.String())), image)
		require.NoError(err)

		err = internal.Cache.Clear()
		require.NoError(err)

		ctx := context.Background()
		images, err := internal.ExpandImage(ctx, image, clade.NewBuildTree(), internal.WithTransport(s.Client().Transport))
		require.NoError(err)
		require.Len(images, 2)
		require.ElementsMatch(
			[]string{"1.0.0", "1.0.1"},
			[]string{images[0].From.Tag(), images[1].From.Tag()},
		)
	})

	t.Run("resolves local tags from build tree", func(t *testing.T) {
		require := require.New(t)

		named, err := reference.ParseNamed("cr.io/repo/local")
		require.NoError(err)

		image := &clade.Image{Named: image_name}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s
  tag: ( tags | semver )`, named.String())), image)
		require.NoError(err)

		bt := clade.NewBuildTree()
		bt.TagsByName["cr.io/repo/local"] = []string{"1.0.42"}

		err = internal.Cache.Clear()
		require.NoError(err)

		ctx := context.Background()
		images, err := internal.ExpandImage(ctx, image, bt, internal.WithTransport(s.Client().Transport))
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("42", images[0].Tags[0])
		require.Equal("1.0.42", images[0].From.Tag())
	})

	t.Run("fails if", func(t *testing.T) {
		tcs := []struct {
			desc string
			port string
			msgs []string
		}{
			{
				desc: "remote repo not exists",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/not-exists
  tag: ( tags | semver)`, u.Host),
				msgs: []string{"get", "tags"},
			},
			{
				desc: "from pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: ( awesome )`, u.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "from pipeline result is not string",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: ( pass 42 )`, u.Host),
				msgs: []string{"convert", "string"},
			},
			{
				desc: "from pipeline results invalid tag format",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: ( log "invalid tag" )`, u.Host),
				msgs: []string{"invalid", "tag"},
			},
			{
				desc: "tag pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [( awesome )]
from:
  name: %s/repo/single
  tag: "1.0.0"`, u.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "tag pipeline results multiple value",
				port: fmt.Sprintf(`
tags: [( log "foo" "bar" )]
from:
  name: %s/repo/single
  tag: "1.0.0"`, u.Host),
				msgs: []string{"size 1", "2"},
			},
			{
				desc: "tag pipeline results type not string",
				port: fmt.Sprintf(`
tags: [( pass 42 )]
from:
  name: %s/repo/single
  tag: "1.0.0"`, u.Host),
				msgs: []string{"string"},
			},
			{
				desc: "tag is duplicated",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/patched
  tag: ( tags | semver )`, u.Host),
				msgs: []string{"duplicated", "foo"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				image := &clade.Image{Named: image_name}
				err = yaml.Unmarshal([]byte(tc.port), image)
				require.NoError(err)

				err = internal.Cache.Clear()
				require.NoError(err)

				ctx := context.Background()
				_, err := internal.ExpandImage(ctx, image, clade.NewBuildTree(), internal.WithTransport(s.Client().Transport))
				require.Error(err)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

func TestLoadBuildTreeFromPorts(t *testing.T) {
	require := require.New(t)

	ports_dir := t.TempDir()
	port_foo_dir := filepath.Join(ports_dir, "foo")
	err := os.Mkdir(port_foo_dir, 0755)
	require.NoError(err)

	empty_manifest := manifest{
		contentType: "application/vnd.docker.distribution.manifest.list.v2+json",
		blob:        []byte{},
	}

	reg := newRegistry(t)
	reg.repos = map[string]*repository{
		"repo/foo": {
			name: "repo/foo",
			manifests: map[string]manifest{
				"1.0.0": empty_manifest,
			},
		},
	}

	s := httptest.NewTLSServer(reg.handler())
	defer s.Close()

	u, err := url.Parse(s.URL)
	require.NoError(err)

	err = os.WriteFile(filepath.Join(port_foo_dir, "port.yaml"), []byte(fmt.Sprintf(`
name: cr.io/repo/foo
images:
  - tags: ["1.0.0"]
    from:
      name: %s/repo/single
      tag: "1.0.0"`, u.Host)), 0644)
	require.NoError(err)

	ctx := context.Background()
	bt := clade.NewBuildTree()
	err = internal.LoadBuildTreeFromPorts(ctx, bt, ports_dir)
	require.NoError(err)
	require.Len(bt.Tree, 2)

	names := []string{}
	bt.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
		names = append(names, name)
		return nil
	})
	require.Equal(names, []string{fmt.Sprintf("%s/repo/single:1.0.0", u.Host), "cr.io/repo/foo:1.0.0"})
}

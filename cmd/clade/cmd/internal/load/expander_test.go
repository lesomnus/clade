package load_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestExpand(t *testing.T) {
	new_image := func(named reference.Named) *clade.Image {
		return &clade.Image{
			Named: named,
			Skip:  new(bool),
		}
	}

	empty_manifest := registry.Manifest{
		ContentType: "application/vnd.docker.distribution.manifest.list.v2+json",
		Blob:        []byte{},
	}

	reg := registry.NewRegistry(t)
	reg.Repos = map[string]*registry.Repository{
		"repo/single": {
			Name: "repo/single",
			Manifests: map[string]registry.Manifest{
				"1.0.0": empty_manifest,
			},
		},
		"repo/patched": {
			Name: "repo/patched",
			Manifests: map[string]registry.Manifest{
				"1.0.0": empty_manifest,
				"1.0.1": empty_manifest,
			},
		},
	}

	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_client := client.NewDistRegistry()
	reg_client.Transport = s.Client().Transport
	reg_client.Cache = &cache.NullCacheStore{}

	expander := load.Expander{
		Registry: reg_client,
	}

	reg_url, err := url.Parse(s.URL)
	require.NoError(t, err)

	image_name, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	t.Run("from static tag", func(t *testing.T) {
		require := require.New(t)

		image := new_image(image_name)
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: "1.0.0"`, reg_url.Host)), image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, image, clade.NewBuildTree())
		require.NoError(err)
		require.Len(images[0].Tags, 1)
		require.Equal("foo", images[0].Tags[0])
		require.Equal("1.0.0", images[0].From.Tag())
	})

	t.Run("executes pipeline with remote tags", func(t *testing.T) {
		require := require.New(t)

		image := new_image(image_name)
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Major )]
from:
  name: %s/repo/single
  tag: ( tags | semver )
args:
  MAJOR: ( printf "%%d" $.Major )`, reg_url.Host)), image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, image, clade.NewBuildTree())
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("1", images[0].Tags[0])
		require.Equal("1.0.0", images[0].From.Tag())
		require.Contains(images[0].Args, "MAJOR")
		require.Equal("1", images[0].Args["MAJOR"])
	})

	t.Run("executes pipeline with local tags from build tree", func(t *testing.T) {
		require := require.New(t)

		local_named, err := reference.ParseNamed("cr.io/repo/local")
		require.NoError(err)

		image := new_image(image_name)
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s
  tag: ( tags | semver )`, local_named.String())), image)
		require.NoError(err)

		bt := clade.NewBuildTree()
		bt.TagsByName[local_named.String()] = []string{"1.0.42"}

		ctx := context.Background()
		images, err := expander.Expand(ctx, image, bt)
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("42", images[0].Tags[0])
		require.Equal("1.0.42", images[0].From.Tag())
	})

	t.Run("expands images as many as remote tags", func(t *testing.T) {
		require := require.New(t)

		image := new_image(image_name)
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s/repo/patched
  tag: ( tags | semver )`, reg_url.Host)), image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, image, clade.NewBuildTree())
		require.NoError(err)
		require.Len(images, 2)
		require.ElementsMatch(
			[]string{"1.0.0", "1.0.1"},
			[]string{images[0].From.Tag(), images[1].From.Tag()},
		)
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
  tag: ( tags | semver)`, reg_url.Host),
				msgs: []string{"get", "tags"},
			},
			{
				desc: "from pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: ( awesome )`, reg_url.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "from pipeline result is not string",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: ( pass 42 )`, reg_url.Host),
				msgs: []string{"convert", "string"},
			},
			{
				desc: "from pipeline results invalid tag format",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/single
  tag: ( log "invalid tag" )`, reg_url.Host),
				msgs: []string{"invalid", "tag"},
			},
			{
				desc: "tag pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [( awesome )]
from:
  name: %s/repo/single
  tag: "1.0.0"`, reg_url.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "tag pipeline results multiple value",
				port: fmt.Sprintf(`
tags: [( log "foo" "bar" )]
from:
  name: %s/repo/single
  tag: "1.0.0"`, reg_url.Host),
				msgs: []string{"size 1", "2"},
			},
			{
				desc: "tag pipeline results type not string",
				port: fmt.Sprintf(`
tags: [( pass 42 )]
from:
  name: %s/repo/single
  tag: "1.0.0"`, reg_url.Host),
				msgs: []string{"string"},
			},
			{
				desc: "tag is duplicated",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/patched
  tag: ( tags | semver )`, reg_url.Host),
				msgs: []string{"duplicated", "foo"},
			},
			{
				desc: "arg pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/single
  tag: "1.0.0"
args:
  FOO: ( awesome )`, reg_url.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "arg pipeline results multiple value",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/single
  tag: "1.0.0"
args:
  FOO: ( log "foo" "bar" )`, reg_url.Host),
				msgs: []string{"size 1", "2"},
			},
			{
				desc: "arg pipeline results type not string",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/single
  tag: "1.0.0"
args:
  FOO: ( pass 42 )`, reg_url.Host),
				msgs: []string{"string"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				image := new_image(image_name)
				err = yaml.Unmarshal([]byte(tc.port), image)
				require.NoError(err)

				ctx := context.Background()
				_, err := expander.Expand(ctx, image, clade.NewBuildTree())
				require.Error(err)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

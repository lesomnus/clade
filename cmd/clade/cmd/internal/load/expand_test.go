package load_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/load"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestExpand_(t *testing.T) {
	reg := registry.NewRegistry()
	srv := registry.NewServer(t, reg)
	s := httptest.NewTLSServer(srv.Handler())
	defer s.Close()

	reg_url, err := url.Parse(s.URL)
	require.NoError(t, err)

	ref_foo, err := reference.ParseNamed(reg_url.Host + "/repo/foo")
	require.NoError(t, err)
	ref_bar, err := reference.ParseNamed(reg_url.Host + "/repo/bar")
	require.NoError(t, err)
	ref_baz, err := reference.ParseNamed(reg_url.Host + "/repo/baz")
	require.NoError(t, err)

	repo_foo := reg.NewRepository(ref_foo)
	repo_bar := reg.NewRepository(ref_bar)
	repo_baz := reg.NewRepository(ref_baz)

	repo_foo.PopulateImageWithTag("1.2.3")
	repo_bar.PopulateImageWithTag("2.3.4")
	repo_baz.PopulateImageWithTag("2.3.4")
	repo_baz.PopulateImageWithTag("2.3.5")

	reg_client := client.NewClient()
	reg_client.Transport = s.Client().Transport

	expander := load.Expand{
		Registry: reg_client,
	}

	t.Run("from static tag", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [first]
from:
  name: %s/repo/foo
  tags: "1.2.3"`, reg_url.Host)), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, &image, clade.NewBuildGraph())
		require.NoError(err)
		require.Len(images[0].Tags, 1)
		require.Equal("first", images[0].Tags[0])
		require.Equal("1.2.3", images[0].From.Primary.Tag())
	})

	t.Run("executes pipeline with remote tags", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Major )]
from:
  name: %s/repo/foo
  tags: ( tags | semver )
args:
  MAJOR: ( printf "%%d" $.Major )`, reg_url.Host)), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, &image, clade.NewBuildGraph())
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("1", images[0].Tags[0])
		require.Equal("1.2.3", images[0].From.Primary.Tag())
		require.Contains(images[0].Args, "MAJOR")
		require.Equal("1", images[0].Args["MAJOR"])
	})

	t.Run("executes pipeline with local tags from build tree", func(t *testing.T) {
		require := require.New(t)

		local_named, err := reference.ParseNamed("cr.io/repo/local")
		require.NoError(err)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s
  tags: ( tags | semver )`, local_named.String())), &image)
		require.NoError(err)

		origin_named, err := reference.Parse("cr.io/origin/name:tag")
		require.NoError(err)

		bg := clade.NewBuildGraph()
		bg.Put(&clade.ResolvedImage{
			Named: local_named,
			Tags:  []string{"1.2.42"},
			From: &clade.ResolvedBaseImage{
				Primary: clade.ResolvedImageReference{
					NamedTagged: origin_named.(reference.NamedTagged),
				},
			},
		})

		ctx := context.Background()
		images, err := expander.Expand(ctx, &image, bg)
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("42", images[0].Tags[0])
		require.Equal("1.2.42", images[0].From.Primary.Tag())
	})

	t.Run("get tags of specific remote repository", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Major )]
from:
  name: %s/repo/foo
  tags: ( tagsOf "%s/repo/bar" | semver )`, reg_url.Host, reg_url.Host)), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, &image, clade.NewBuildGraph())
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("2", images[0].Tags[0])
		require.Equal("2.3.4", images[0].From.Primary.Tag())
	})

	t.Run("get tags of specific local repository", func(t *testing.T) {
		require := require.New(t)

		local_named, err := reference.ParseNamed("cr.io/repo/local")
		require.NoError(err)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s/repo/foo
  tags: ( tagsOf "%s" | semver )`, reg_url.Host, local_named.String())), &image)
		require.NoError(err)

		origin_named, err := reference.Parse("cr.io/origin/name:tag")
		require.NoError(err)

		bg := clade.NewBuildGraph()
		bg.Put(&clade.ResolvedImage{
			Named: local_named,
			Tags:  []string{"1.2.42"},
			From: &clade.ResolvedBaseImage{
				Primary: clade.ResolvedImageReference{
					NamedTagged: origin_named.(reference.NamedTagged),
				},
			},
		})

		ctx := context.Background()
		images, err := expander.Expand(ctx, &image, bg)
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("42", images[0].Tags[0])
		require.Equal("1.2.42", images[0].From.Primary.Tag())
	})

	t.Run("expands images as many as remote tags", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_baz}
		err = yaml.Unmarshal([]byte(fmt.Sprintf(`
tags: [( printf "%%d" $.Patch )]
from:
  name: %s/repo/baz
  tags: ( tags | semver )`, reg_url.Host)), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := expander.Expand(ctx, &image, clade.NewBuildGraph())
		require.NoError(err)
		require.Len(images, 2)
		require.ElementsMatch(
			[]string{"2.3.4", "2.3.5"},
			[]string{images[0].From.Primary.Tag(), images[1].From.Primary.Tag()},
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
  tags: ( tags | semver)`, reg_url.Host),
				msgs: []string{"get", "tags"},
			},
			{
				desc: "invalid repo format",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/foo
  tags: ( tagsOf "invalid repo"  | semver)`, reg_url.Host),
				msgs: []string{"invalid", "format"},
			},
			{
				desc: "from pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/foo
  tags: ( awesome )`, reg_url.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "from pipeline result is not string",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/foo
  tags: ( pass 42 )`, reg_url.Host),
				msgs: []string{"result", "string"},
			},
			{
				desc: "from pipeline results invalid tag format",
				port: fmt.Sprintf(`
tags: [foo]
from:
  name: %s/repo/foo
  tags: ( log "invalid tag" )`, reg_url.Host),
				msgs: []string{"invalid", "tag"},
			},
			{
				desc: "tag pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [( awesome )]
from:
  name: %s/repo/foo
  tags: "1.0.0"`, reg_url.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "tag pipeline results multiple value",
				port: fmt.Sprintf(`
tags: [( log "foo" "bar" )]
from:
  name: %s/repo/foo
  tags: "1.0.0"`, reg_url.Host),
				msgs: []string{"sized 1", "2"},
			},
			{
				desc: "tag pipeline results type not string",
				port: fmt.Sprintf(`
tags: [( pass 42 )]
from:
  name: %s/repo/foo
  tags: "1.0.0"`, reg_url.Host),
				msgs: []string{"string"},
			},
			{
				desc: "tag is duplicated",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/baz
  tags: ( tags | semver )`, reg_url.Host),
				msgs: []string{"duplicated", "foo"},
			},
			{
				desc: "arg pipeline with undefined functions",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/foo
  tags: "1.0.0"
args:
  FOO: ( awesome )`, reg_url.Host),
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "arg pipeline results multiple value",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/foo
  tags: "1.0.0"
args:
  FOO: ( log "foo" "bar" )`, reg_url.Host),
				msgs: []string{"sized 1", "2"},
			},
			{
				desc: "arg pipeline results type not string",
				port: fmt.Sprintf(`
tags: [ foo ]
from:
  name: %s/repo/foo
  tags: "1.0.0"
args:
  FOO: ( pass 42 )`, reg_url.Host),
				msgs: []string{"string"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				image := clade.Image{Named: ref_foo}
				err = yaml.Unmarshal([]byte(tc.port), &image)
				require.NoError(err)

				ctx := context.Background()
				_, err := expander.Expand(ctx, &image, clade.NewBuildGraph())
				require.Error(err)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

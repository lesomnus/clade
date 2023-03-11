package clade_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/internal/registry"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

func TestPortLoaderLoad(t *testing.T) {
	reg := registry.NewRegistry()
	origin_foo := reg.NewRepository(must(reference.ParseNamed("cr.io/origin/foo")))
	origin_bar := reg.NewRepository(must(reference.ParseNamed("cr.io/origin/bar")))
	origin_baz := reg.NewRepository(must(reference.ParseNamed("cr.io/origin/baz")))

	origin_foo.PopulateImageWithTag("foo")
	origin_bar.PopulateImageWithTag("bar")
	origin_baz.PopulateImageWithTag("baz")

	make_port := func(port string) *clade.Port {
		rst := &clade.Port{}
		err := yaml.Unmarshal([]byte(port), &rst)
		require.NoError(t, err)

		return rst
	}

	t.Run("static", func(t *testing.T) {
		//              origin/bar:bar -| repo/bar:c
		//                              | repo/bar:d
		//                             /            \
		// origin/foo:foo -| repo/foo:a              \
		//                 | repo/foo:b --------------| repo/baz:e
		//                                           /
		//                             origin/baz:baz

		require := require.New(t)

		ports := []*clade.Port{
			make_port(`
name: cr.io/repo/foo
images:
  - tags: [a, b]
    from:
      name: cr.io/origin/foo
      tags: foo`),
			make_port(`
name: cr.io/repo/bar
images:
  - tags: [c, d]
    from:
      name: cr.io/origin/bar
      tags: bar
      with:
        - cr.io/repo/foo:a`),
			make_port(`
name: cr.io/repo/baz
images:
  - tags: [e]
    from:
      name: cr.io/origin/baz
      tags: baz
      with:
        - cr.io/repo/foo:b
        - cr.io/repo/bar:d`),
		}

		port_loader := clade.PortLoader{
			Registry: reg,
		}

		ctx := context.Background()
		bg := clade.NewBuildGraph()
		err := port_loader.Load(ctx, bg, ports)
		require.NoError(err)

		snapshot := bg.Snapshot()
		require.Len(snapshot, 6)
		require.Contains(snapshot, "cr.io/origin/foo:foo")
		require.Contains(snapshot, "cr.io/origin/bar:bar")
		require.Contains(snapshot, "cr.io/origin/baz:baz")
		require.Contains(snapshot, "cr.io/repo/foo:a")
		require.Contains(snapshot, "cr.io/repo/bar:c")
		require.Contains(snapshot, "cr.io/repo/baz:e")

		entry_foo := snapshot["cr.io/repo/foo:a"]
		entry_bar := snapshot["cr.io/repo/bar:c"]
		entry_baz := snapshot["cr.io/repo/baz:e"]

		require.Equal(uint(1), entry_foo.Level)
		require.Equal(uint(2), entry_bar.Level)
		require.Equal(uint(3), entry_baz.Level)

		require.Subset(maps.Keys(entry_foo.Group), maps.Keys(entry_bar.Group))
		require.Subset(maps.Keys(entry_foo.Group), maps.Keys(entry_baz.Group))
		require.Subset(maps.Keys(entry_bar.Group), maps.Keys(entry_baz.Group))

		require.NotSubset(maps.Keys(entry_baz.Group), maps.Keys(entry_bar.Group))
		require.NotSubset(maps.Keys(entry_baz.Group), maps.Keys(entry_foo.Group))

		require.ElementsMatch(maps.Keys(entry_bar.Group), maps.Keys(entry_foo.Group))
	})
}

func TestPortLoaderExpand(t *testing.T) {
	reg := registry.NewRegistry()

	ref_foo, err := reference.ParseNamed("cr.io/repo/foo")
	require.NoError(t, err)
	ref_bar, err := reference.ParseNamed("cr.io/repo/bar")
	require.NoError(t, err)
	ref_baz, err := reference.ParseNamed("cr.io/repo/baz")
	require.NoError(t, err)

	repo_foo := reg.NewRepository(ref_foo)
	repo_bar := reg.NewRepository(ref_bar)
	repo_baz := reg.NewRepository(ref_baz)

	repo_foo.PopulateImageWithTag("1.2.3")
	repo_bar.PopulateImageWithTag("2.3.4")
	repo_baz.PopulateImageWithTag("2.3.4")
	repo_baz.PopulateImageWithTag("2.3.5")

	port_loader := clade.PortLoader{
		Registry: reg,
	}

	t.Run("from static tag", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(`
tags: [first]
from:
  name: cr.io/repo/foo
  tags: "1.2.3"`), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := port_loader.Expand(ctx, &image, clade.NewBuildGraph())
		require.NoError(err)
		require.Len(images[0].Tags, 1)
		require.Equal("first", images[0].Tags[0])
		require.Equal("1.2.3", images[0].From.Primary.Tag())
	})

	t.Run("executes pipeline with remote tags", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(`
tags: [( printf "%d" $.Major )]
from:
  name: cr.io/repo/foo
  tags: ( tags | semver )
args:
  MAJOR: ( printf "%d" $.Major )`), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := port_loader.Expand(ctx, &image, clade.NewBuildGraph())
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
		images, err := port_loader.Expand(ctx, &image, bg)
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("42", images[0].Tags[0])
		require.Equal("1.2.42", images[0].From.Primary.Tag())
	})

	t.Run("get tags of specific remote repository", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_foo}
		err = yaml.Unmarshal([]byte(`
tags: [( printf "%d" $.Major )]
from:
  name: cr.io/repo/foo
  tags: ( tagsOf "cr.io/repo/bar" | semver )`), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := port_loader.Expand(ctx, &image, clade.NewBuildGraph())
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
  name: cr.io/repo/foo
  tags: ( tagsOf "%s" | semver )`, local_named.String())), &image)
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
		images, err := port_loader.Expand(ctx, &image, bg)
		require.NoError(err)
		require.Len(images, 1)
		require.Len(images[0].Tags, 1)
		require.Equal("42", images[0].Tags[0])
		require.Equal("1.2.42", images[0].From.Primary.Tag())
	})

	t.Run("expands images as many as remote tags", func(t *testing.T) {
		require := require.New(t)

		image := clade.Image{Named: ref_baz}
		err = yaml.Unmarshal([]byte(`
tags: [( printf "%d" $.Patch )]
from:
  name: cr.io/repo/baz
  tags: ( tags | semver )`), &image)
		require.NoError(err)

		ctx := context.Background()
		images, err := port_loader.Expand(ctx, &image, clade.NewBuildGraph())
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
				port: `
tags: [foo]
from:
  name: cr.io/repo/not-exists
  tags: ( tags | semver )`,
				msgs: []string{"name", "unknown"},
			},
			{
				desc: "invalid repo format",
				port: `
tags: [foo]
from:
  name: cr.io/repo/foo
  tags: ( tagsOf "invalid repo"  | semver )`,
				msgs: []string{"invalid", "format"},
			},
			{
				desc: "from pipeline with undefined functions",
				port: `
tags: [foo]
from:
  name: cr.io/repo/foo
  tags: ( awesome )`,
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "from pipeline result is not string",
				port: `
tags: [foo]
from:
  name: cr.io/repo/foo
  tags: ( pass 42 )`,
				msgs: []string{"result", "string"},
			},
			{
				desc: "from pipeline results invalid tag format",
				port: `
tags: [foo]
from:
  name: cr.io/repo/foo
  tags: ( log "invalid tag" )`,
				msgs: []string{"invalid", "tag"},
			},
			{
				desc: "tag pipeline with undefined functions",
				port: `
tags: [( awesome )]
from:
  name: cr.io/repo/foo
  tags: "1.0.0"`,
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "tag pipeline results multiple value",
				port: `
tags: [( log "foo" "bar" )]
from:
  name: cr.io/repo/foo
  tags: "1.0.0"`,
				msgs: []string{"sized 1", "2"},
			},
			{
				desc: "tag pipeline results type not string",
				port: `
tags: [( pass 42 )]
from:
  name: cr.io/repo/foo
  tags: "1.0.0"`,
				msgs: []string{"string"},
			},
			{
				desc: "tag is duplicated",
				port: `
tags: [ foo ]
from:
  name: cr.io/repo/baz
  tags: ( tags | semver )`,
				msgs: []string{"duplicated", "foo"},
			},
			{
				desc: "arg pipeline with undefined functions",
				port: `
tags: [ foo ]
from:
  name: cr.io/repo/foo
  tags: "1.0.0"
args:
  FOO: ( awesome )`,
				msgs: []string{"awesome", "defined"},
			},
			{
				desc: "arg pipeline results multiple value",
				port: `
tags: [ foo ]
from:
  name: cr.io/repo/foo
  tags: "1.0.0"
args:
  FOO: ( log "foo" "bar" )`,
				msgs: []string{"sized 1", "2"},
			},
			{
				desc: "arg pipeline results type not string",
				port: `
tags: [ foo ]
from:
  name: cr.io/repo/foo
  tags: "1.0.0"
args:
  FOO: ( pass 42 )`,
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
				_, err := port_loader.Expand(ctx, &image, clade.NewBuildGraph())
				require.Error(err)
				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

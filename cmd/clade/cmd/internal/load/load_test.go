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
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

func must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func TestExapndLoad(t *testing.T) {
	reg := registry.NewRegistry()
	origin_foo := reg.NewRepository(must(reference.WithName("origin/foo")))
	origin_bar := reg.NewRepository(must(reference.WithName("origin/bar")))
	origin_baz := reg.NewRepository(must(reference.WithName("origin/baz")))

	origin_foo.PopulateImageWithTag("foo")
	origin_bar.PopulateImageWithTag("bar")
	origin_baz.PopulateImageWithTag("baz")

	srv := registry.NewServer(t, reg)
	s := httptest.NewTLSServer(srv.Handler())
	defer s.Close()

	reg_url, err := url.Parse(s.URL)
	require.NoError(t, err)

	make_port := func(port string) *clade.Port {
		rst := &clade.Port{}
		err := yaml.Unmarshal([]byte(fmt.Sprintf(port, reg_url.Host)), &rst)
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
      name: %s/origin/foo
      tags: foo`),
			make_port(`
name: cr.io/repo/bar
images:
  - tags: [c, d]
    from:
      name: %s/origin/bar
      tags: bar
      with:
        - cr.io/repo/foo:a`),
			make_port(`
name: cr.io/repo/baz
images:
  - tags: [e]
    from:
      name: %s/origin/baz
      tags: baz
      with:
        - cr.io/repo/foo:b
        - cr.io/repo/bar:d`),
		}

		reg_client := client.NewClient()
		reg_client.Transport = s.Client().Transport

		expander := load.Expand{
			Registry: reg_client,
		}

		ctx := context.Background()
		bg := clade.NewBuildGraph()
		err = expander.Load(ctx, bg, ports)
		require.NoError(err)

		snapshot := bg.Snapshot()
		require.Len(snapshot, 3)
		require.Contains(snapshot, "cr.io/repo/foo")
		require.Contains(snapshot, "cr.io/repo/bar")
		require.Contains(snapshot, "cr.io/repo/baz")

		entry_foo := snapshot["cr.io/repo/foo"]
		entry_bar := snapshot["cr.io/repo/bar"]
		entry_baz := snapshot["cr.io/repo/baz"]

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

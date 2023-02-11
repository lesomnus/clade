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
	"github.com/lesomnus/clade/tree"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestLoad(t *testing.T) {
	require := require.New(t)

	ref_foo, err := reference.WithName("repo/foo")
	require.NoError(err)

	repo_foo := registry.NewRepository(ref_foo)
	repo_foo.PopulateImageWithTag("1.0.0")

	reg := registry.NewRegistry()
	reg.Repos[ref_foo.Name()] = repo_foo

	srv := registry.NewServer(t, reg)
	s := httptest.NewTLSServer(srv.Handler())
	defer s.Close()

	reg_url, err := url.Parse(s.URL)
	require.NoError(err)

	var port clade.Port
	err = yaml.Unmarshal([]byte(fmt.Sprintf(`
name: cr.io/repo/foo
images:
  - tags: ["1.0.0"]
    from:
      name: %s/repo/foo
      tag: "1.0.0"`, reg_url.Host)), &port)
	require.NoError(err)

	reg_client := client.NewRegistry()
	reg_client.Transport = s.Client().Transport
	reg_client.Cache.Tags = &cache.NullTagCache{}

	expander := load.Expander{
		Registry: reg_client,
	}

	loader := load.NewLoader()
	loader.Expander = expander

	ctx := context.Background()
	bt := clade.NewBuildTree()
	err = loader.Load(ctx, bt, []*clade.Port{&port})
	require.NoError(err)
	require.NoError(err)
	require.Len(bt.Tree, 2)

	names := []string{}
	bt.Walk(func(level int, name string, node *tree.Node[*clade.ResolvedImage]) error {
		names = append(names, name)
		return nil
	})
	require.Equal([]string{fmt.Sprintf("%s/repo/foo:1.0.0", reg_url.Host), "cr.io/repo/foo:1.0.0"}, names)
}

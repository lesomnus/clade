package load_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

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

	empty_manifest := registry.Manifest{
		ContentType: "application/vnd.docker.distribution.manifest.list.v2+json",
		Blob:        []byte{},
	}

	reg := registry.NewRegistry(t)
	reg.Repos = map[string]*registry.Repository{
		"repo/foo": {
			Name: "repo/foo",
			Manifests: map[string]registry.Manifest{
				"1.0.0": empty_manifest,
			},
		},
	}

	s := httptest.NewTLSServer(reg.Handler())
	defer s.Close()

	reg_url, err := url.Parse(s.URL)
	require.NoError(err)

	var port clade.Port
	err = yaml.Unmarshal([]byte(fmt.Sprintf(`
name: cr.io/repo/foo
images:
  - tags: ["1.0.0"]
    from:
      name: %s/repo/single
      tag: "1.0.0"`, reg_url.Host)), &port)
	require.NoError(err)

	reg_client := client.NewDistRegistry()
	reg_client.Transport = s.Client().Transport
	reg_client.Cache = &cache.NullCacheStore{}

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
	require.Equal(names, []string{fmt.Sprintf("%s/repo/single:1.0.0", reg_url.Host), "cr.io/repo/foo:1.0.0"})
}

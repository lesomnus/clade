package internal_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/distribution/distribution/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/stretchr/testify/require"
)

func TestRepositoryTagsAll(t *testing.T) {
	require := require.New(t)

	reg := newRegistry(t)
	reg.repos["repo/name"] = repository{
		tags: []string{"a", "b", "c"},
	}

	s := httptest.NewTLSServer(reg.handler())
	defer s.Close()

	u, err := url.Parse(s.URL)
	require.NoError(err)

	named, err := reference.ParseNamed(fmt.Sprintf("%s/repo/name", u.Host))
	require.NoError(err)

	repo, err := internal.NewRepositoryWithRoundTripper(s.Client().Transport, named)
	require.NoError(err)

	ctx := context.Background()
	tags, err := repo.Tags(ctx).All(ctx)
	require.NoError(err)

	require.ElementsMatch(reg.repos["repo/name"].tags, tags)
}

package cache_test

import (
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestRepositoryNamed(t *testing.T) {
	require := require.New(t)

	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(err)

	tagged, err := reference.WithTag(named, "foo")
	require.NoError(err)

	tmp := t.TempDir()
	reg := cache.NewRegistry(tmp)
	repo, err := reg.Repository(tagged)
	require.NoError(err)

	name := repo.Named().String()
	require.NotContains(name, "foo")
}

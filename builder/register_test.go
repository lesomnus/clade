package builder_test

import (
	"testing"

	"github.com/lesomnus/clade/builder"
	"github.com/stretchr/testify/require"
)

func TestRegister(t *testing.T) {
	require := require.New(t)

	is_invoked := false
	builder.Register("test", func(conf builder.BuilderConfig) (builder.Builder, error) {
		is_invoked = true
		return nil, nil
	})

	_, err := builder.New("test", builder.BuilderConfig{})
	require.NoError(err)
	require.True(is_invoked)
}

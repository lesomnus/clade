package plf_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/lesomnus/clade/plf"
	"github.com/lesomnus/clade/sv"
	"github.com/stretchr/testify/require"
)

func TestFuncs(t *testing.T) {
	require := require.New(t)

	m1 := plf.Funcs()
	m2 := plf.Funcs()

	for k, v1 := range m1 {
		v2, ok := m2[k]
		require.True(ok)
		require.Equal(fmt.Sprintf("%p", v1), fmt.Sprintf("%p", v2))
	}

	m1["foo"] = "foo"
	require.NotContains(m2, "foo")
}

func TestConvs(t *testing.T) {
	require := require.New(t)

	m1 := plf.Convs()
	m2 := plf.Convs()

	for k, v1 := range m1 {
		v2, ok := m2[k]
		require.True(ok)
		for k, c1 := range v1 {
			c2, ok := v2[k]
			require.True(ok)
			require.Equal(fmt.Sprintf("%p", c1), fmt.Sprintf("%p", c2))
		}
	}

	m1[reflect.TypeOf(true)] = make(map[reflect.Type]func(v reflect.Value) (any, error))
	require.NotContains(m2, "foo")

	_, err := m1.ConvertTo(reflect.TypeOf(&sv.Version{}), "3.1.4")
	require.NoError(err)
}

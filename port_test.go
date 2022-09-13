package clade_test

import (
	"testing"

	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestDeduplicateBySemver(t *testing.T) {
	type Pair struct {
		lhs []string
		rhs []string
	}

	tcs := []struct {
		given    Pair
		expected Pair
	}{
		{
			given: Pair{
				lhs: []string{"1.0.0", "1.0", "1"},
				rhs: []string{"1.0.0", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.0.0", "1.0", "1"},
				rhs: []string{},
			},
		},
		{
			given: Pair{
				lhs: []string{"1.0.1", "1.0", "1"},
				rhs: []string{"1.0.0", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.0.1", "1.0", "1"},
				rhs: []string{"1.0.0"},
			},
		},
		{
			given: Pair{
				lhs: []string{"1.0.0", "1.0", "1"},
				rhs: []string{"1.0.1", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.0.0"},
				rhs: []string{"1.0.1", "1.0", "1"},
			},
		},
		{
			given: Pair{
				lhs: []string{"1.1.1", "1.1", "1"},
				rhs: []string{"1.0.2", "1.0", "1"},
			},
			expected: Pair{
				lhs: []string{"1.1.1", "1.1", "1"},
				rhs: []string{"1.0.2", "1.0"},
			},
		},
	}
	for _, tc := range tcs {
		actual := Pair{
			lhs: slices.Clone(tc.given.lhs),
			rhs: slices.Clone(tc.given.rhs),
		}

		err := clade.DeduplicateBySemver(&actual.lhs, &actual.rhs)
		require.NoError(t, err)
		require.ElementsMatch(t, tc.expected.lhs, actual.lhs)
		require.ElementsMatch(t, tc.expected.rhs, actual.rhs)
	}
}

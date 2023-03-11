package clade_test

import (
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/clade"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestNewBuildPlanIterations(t *testing.T) {
	tcs := []struct {
		desc string
		bg   *clade.BuildGraph

		expected [][][]string
	}{
		{
			desc: "single",
			bg: func() *clade.BuildGraph {
				bg := clade.NewBuildGraph()
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/foo")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/origin/foo:1")).(reference.NamedTagged),
						},
					},
				})
				return bg
			}(),
			expected: [][][]string{
				{
					{"cr.io/repo/foo:1"},
				},
			},
		},
		{
			desc: "collect",
			bg: func() *clade.BuildGraph {
				bg := clade.NewBuildGraph()
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/foo")),
					Tags:  []string{"1", "2"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/origin/foo:1")).(reference.NamedTagged),
						},
					},
				})
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/bar")),
					Tags:  []string{"1", "2"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/origin/bar:1")).(reference.NamedTagged),
						},
					},
				})
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/baz")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/repo/foo:2")).(reference.NamedTagged),
						},
						Secondaries: []clade.ResolvedImageReference{
							{NamedTagged: must(reference.ParseNamed("cr.io/repo/bar:2")).(reference.NamedTagged)},
						},
					},
				})
				return bg
			}(),
			expected: [][][]string{
				{
					{"cr.io/repo/foo:1", "cr.io/repo/bar:1"},
				},
				{
					{"cr.io/repo/baz:1"},
				},
			},
		},
		{
			desc: "same base",
			bg: func() *clade.BuildGraph {
				bg := clade.NewBuildGraph()
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/foo")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/origin/foo:1")).(reference.NamedTagged),
						},
					},
				})
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/bar")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/repo/foo:1")).(reference.NamedTagged),
						},
					},
				})
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/baz")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/repo/foo:1")).(reference.NamedTagged),
						},
					},
				})
				return bg
			}(),
			expected: [][][]string{
				{
					{"cr.io/repo/foo:1"},
				},
				{
					{"cr.io/repo/bar:1"},
					{"cr.io/repo/baz:1"},
				},
			},
		},
		{
			desc: "nested base",
			bg: func() *clade.BuildGraph {
				bg := clade.NewBuildGraph()
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/foo")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/origin/foo:1")).(reference.NamedTagged),
						},
					},
				})
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/bar")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/repo/foo:1")).(reference.NamedTagged),
						},
					},
				})
				bg.Put(&clade.ResolvedImage{
					Named: must(reference.ParseNamed("cr.io/repo/baz")),
					Tags:  []string{"1"},
					From: &clade.ResolvedBaseImage{
						Primary: clade.ResolvedImageReference{
							NamedTagged: must(reference.ParseNamed("cr.io/repo/foo:1")).(reference.NamedTagged),
						},
						Secondaries: []clade.ResolvedImageReference{
							{NamedTagged: must(reference.ParseNamed("cr.io/repo/bar:1")).(reference.NamedTagged)},
						},
					},
				})
				return bg
			}(),
			expected: [][][]string{
				{
					{"cr.io/repo/foo:1"},
				},
				{
					{"cr.io/repo/bar:1"},
				},
				{
					{"cr.io/repo/baz:1"},
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)
			plan := clade.NewBuildPlan(tc.bg)

			actual := plan.Iterations
			require.Len(actual, len(tc.expected))
			for level, expected_ref_groups := range tc.expected {
				actual_ref_groups := actual[level]
				require.Len(actual_ref_groups, len(expected_ref_groups))

				for _, expected_ref_group := range expected_ref_groups {
					done := false
					for _, actual_ref_group := range actual_ref_groups {
						if !slices.Contains(actual_ref_group, expected_ref_group[0]) {
							continue
						}

						done = true
						require.ElementsMatch(expected_ref_group, actual_ref_group)
					}
					if !done {
						require.Fail("ref omitted", expected_ref_group[0])
					}
				}
			}
		})
	}
}

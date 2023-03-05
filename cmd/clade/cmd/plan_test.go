package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestPlanCmd(t *testing.T) {
	tcs := []struct {
		desc     string
		args     []string
		prepare  func(t *testing.T) *TmpPortDir
		expected clade.BuildPlan
	}{
		{
			desc: "all images are planed if no arguments given",
			args: []string{},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/origin/bar:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}, {"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
		{
			desc: "only images that depend on the image given as argument are planed",
			args: []string{"cr.io/repo/bar:1"},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/origin/bar:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
		{
			desc: "dependent image is planed if secondary image is given",
			args: []string{"cr.io/repo/bar:1"},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/origin/bar:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from:
      name: cr.io/repo/foo
      tags: 1
      with: [cr.io/repo/bar:1]`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
		{
			desc: "same dependent image does not cause duplicated error",
			args: []string{"cr.io/repo/foo:1", "cr.io/repo/bar:1"},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/origin/bar:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from:
      name: cr.io/repo/foo
      tags: 1
      with: [cr.io/repo/bar:1]`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1", "cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
		{
			desc: "put dependent image does not cause duplicated error",
			args: []string{"cr.io/repo/foo:1", "cr.io/repo/bar:1", "cr.io/repo/baz:1"},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/origin/bar:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}, {"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
		{
			desc: "nested dependency",
			args: []string{"cr.io/repo/foo:1", "cr.io/repo/bar:1", "cr.io/repo/baz:1"},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/repo/foo:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from:
      name: cr.io/repo/foo
      tags: 1
      with: [cr.io/repo/bar:1]`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}},
					{{"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
		{
			desc: `get references from stdin if "-" is given as argument`,
			args: []string{"-", "cr.io/repo/foo:1\n cr.io/repo/bar:1"},
			prepare: func(t *testing.T) *TmpPortDir {
				ports := NewTmpPortDir(t)
				ports.AddRaw("foo", `
name: cr.io/repo/foo
images:
  - tags: [1]
    from: cr.io/origin/foo:1`)
				ports.AddRaw("bar", `
name: cr.io/repo/bar
images:
  - tags: [1]
    from: cr.io/origin/bar:1`)
				ports.AddRaw("baz", `
name: cr.io/repo/baz
images:
  - tags: [1]
    from: cr.io/repo/bar:1`)

				return ports
			},
			expected: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}, {"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			ports := tc.prepare(t)
			args := tc.args

			buff_in := new(bytes.Buffer)
			if len(tc.args) > 1 && tc.args[0] == "-" {
				buff_in.Write([]byte(tc.args[1]))
				args = []string{"-"}
			}

			buff_out := new(bytes.Buffer)
			svc := cmd.NewCmdService()
			svc.In = buff_in
			svc.Out = buff_out
			svc.RegistryClient = registry.NewRegistry()

			flags := cmd.PlanFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: ports.Dir,
				},
			}

			c := cmd.CreatePlanCmd(&flags, svc)
			c.SetOut(os.Stderr)
			c.SetArgs(args)

			err := c.Execute()
			require.NoError(err)

			output := buff_out.Bytes()
			var actual clade.BuildPlan
			err = json.Unmarshal(output, &actual)
			require.NoError(err, string(output))

			require.Len(actual.Iterations, len(tc.expected.Iterations))
			for level, expected_collections := range tc.expected.Iterations {
				actual_collections := actual.Iterations[level]
				require.Len(actual_collections, len(expected_collections))

				for _, expected_collection := range expected_collections {
					done := false
					for _, actual_collection := range actual_collections {
						if !slices.Contains(actual_collection, expected_collection[0]) {
							continue
						}

						done = true
						require.ElementsMatch(expected_collection, actual_collection)
					}
					if !done {
						require.Fail("ref omitted", expected_collection[0])
					}
				}
			}
		})
	}
}

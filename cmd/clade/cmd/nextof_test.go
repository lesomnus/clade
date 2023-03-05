package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/registry"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestNextofCmd(t *testing.T) {
	tcs := []struct {
		desc     string
		args     []string
		plan     clade.BuildPlan
		prepare  func(t *testing.T) *TmpPortDir
		expected [][]string
	}{
		{
			desc: "prints empty array if there is no next",
			args: []string{"cr.io/repo/foo:1"},
			plan: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}, {"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
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
			expected: [][]string{},
		},
		{
			desc: "prints next collection if exists",
			args: []string{"cr.io/repo/bar:1"},
			plan: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}, {"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
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
			expected: [][]string{{"cr.io/repo/baz:1"}},
		},
		{
			desc: "only direct next one is printed",
			args: []string{"cr.io/repo/foo:1"},
			plan: clade.BuildPlan{
				Iterations: [][][]string{
					{{"cr.io/repo/foo:1"}},
					{{"cr.io/repo/bar:1"}},
					{{"cr.io/repo/baz:1"}},
				},
			},
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
			expected: [][]string{{"cr.io/repo/bar:1"}},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			require := require.New(t)

			plan_filepath := filepath.Join(t.TempDir(), "plan.json")
			plan_data, err := json.Marshal(tc.plan)
			require.NoError(err)

			err = os.WriteFile(plan_filepath, plan_data, 0644)
			require.NoError(err)

			ports := tc.prepare(t)
			args := append([]string{plan_filepath}, tc.args...)

			buff_out := new(bytes.Buffer)
			svc := cmd.NewCmdService()
			svc.Out = buff_out
			svc.RegistryClient = registry.NewRegistry()

			flags := cmd.NextofFlags{
				RootFlags: &cmd.RootFlags{
					PortsPath: ports.Dir,
				},
			}

			c := cmd.CreateNextofCmd(&flags, svc)
			c.SetOut(os.Stderr)
			c.SetArgs(args)

			err = c.Execute()
			require.NoError(err)

			output := buff_out.Bytes()
			var actual [][]string
			err = json.Unmarshal(output, &actual)
			require.NoError(err, string(output))
			require.Len(actual, len(tc.expected))
			for _, expected_collection := range tc.expected {
				done := false
				for _, actual_collection := range actual {
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
		})
	}
}

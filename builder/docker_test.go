package builder_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/distribution/distribution/v3/reference"
	"github.com/lesomnus/boolal"
	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/builder"
	"github.com/stretchr/testify/require"
)

func TestNewDockerCmdBuilder(t *testing.T) {
	t.Run("--platform is parsed", func(t *testing.T) {
		require := require.New(t)
		b, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
			DryRun: true,
			Args:   []string{"foo", "--platform", "linux/amd64,linux/arm64", "bar"},
		})
		require.NoError(err)
		require.Equal([]string{"foo", "bar"}, b.Config.Args)
		require.Len(b.Platforms, 2)
		require.ElementsMatch(b.Platforms, []builder.PlatformSpecifier{
			{Os: "linux", Arch: "amd64"},
			{Os: "linux", Arch: "arm64"},
		})
	})

	t.Run("fails if", func(t *testing.T) {
		t.Run("--platform does not have a value", func(t *testing.T) {
			tcs := []struct {
				desc string
				args []string
			}{
				{
					desc: "no next arg",
					args: []string{"--platform"},
				},
				{
					desc: "no value by long flag",
					args: []string{"--platform", "--foo"},
				},
				{
					desc: "no value by short flag",
					args: []string{"--platform", "-foo"},
				},
				{
					desc: "empty value",
					args: []string{"--platform", ""},
				},
			}
			for _, tc := range tcs {
				t.Run(tc.desc, func(t *testing.T) {
					require := require.New(t)
					_, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
						DryRun: true,
						Args:   tc.args,
					})
					require.ErrorContains(err, "--platform")
					require.ErrorContains(err, "value")
				})
			}
		})

		t.Run("invalid format of --platform", func(t *testing.T) {
			tcs := []struct {
				desc  string
				value string
			}{
				{
					desc:  "no arch",
					value: "os/",
				},
				{
					desc:  "no os",
					value: "/arch",
				},
				{
					desc:  "no slash",
					value: "os",
				},
				{
					desc:  "empty among many",
					value: "os/arch,,foo/bar",
				},
				{
					desc:  "invalid among many",
					value: "os/arch,baz,foo/bar",
				},
			}
			for _, tc := range tcs {
				t.Run(tc.desc, func(t *testing.T) {
					require := require.New(t)
					_, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
						DryRun: true,
						Args:   []string{"--platform", tc.value},
					})
					require.ErrorContains(err, "--platform")
					require.ErrorContains(err, "format")
				})
			}
		})
	})
}

func TestDockerCmdBuild(t *testing.T) {
	named, err := reference.ParseNamed("cr.io/repo/name")
	require.NoError(t, err)

	t.Run("deref ID annotation for image", func(t *testing.T) {
		require := require.New(t)

		b, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
			DryRun: true,
			Args:   []string{"--platform", "os/arch"},
		})
		require.NoError(err)

		o := bytes.Buffer{}
		err = b.Build(&clade.ResolvedImage{
			Named:       named,
			Tags:        []string{"tag"},
			ContextPath: ".",
			Platform:    boolal.Or("os", "arch"),
		}, builder.BuildOption{
			DerefId: "foo",
			Stdout:  &o,
			Stderr:  io.Discard,
		})
		require.NoError(err)

		rst := o.String()
		require.Contains(rst, "annotation.")
	})

	t.Run("deref ID annotation for index", func(t *testing.T) {
		require := require.New(t)

		b, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
			DryRun: true,
			Args:   []string{"--platform", "os/arch,os/amd64"},
		})
		require.NoError(err)

		o := bytes.Buffer{}
		err = b.Build(&clade.ResolvedImage{
			Named:       named,
			Tags:        []string{"tag"},
			ContextPath: ".",
			Platform:    boolal.Or("os", "arch"),
		}, builder.BuildOption{
			DerefId: "foo",
			Stdout:  &o,
			Stderr:  io.Discard,
		})
		require.NoError(err)

		rst := o.String()
		require.Contains(rst, "annotation-index.")
	})

	t.Run("value of --platform is filtered", func(t *testing.T) {
		require := require.New(t)

		b, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
			DryRun: true,
			Args:   []string{"--platform", "os/arch,os/amd64"},
		})
		require.NoError(err)

		o := bytes.Buffer{}
		err = b.Build(&clade.ResolvedImage{
			Named:       named,
			Tags:        []string{"tag"},
			ContextPath: ".",
			Platform:    boolal.And("os", "arch"),
		}, builder.BuildOption{
			Stdout: &o,
			Stderr: io.Discard,
		})
		require.NoError(err)

		rst := o.String()
		require.Contains(rst, "--platform os/arch")
	})

	t.Run("multiple tags", func(t *testing.T) {
		require := require.New(t)

		b, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
			DryRun: true,
		})
		require.NoError(err)

		o := bytes.Buffer{}
		err = b.Build(&clade.ResolvedImage{
			Named:       named,
			Tags:        []string{"foo", "bar"},
			ContextPath: ".",
			Platform:    boolal.And("os", "arch"),
		}, builder.BuildOption{
			Stdout: &o,
			Stderr: io.Discard,
		})
		require.NoError(err)

		rst := o.String()
		require.Contains(rst, fmt.Sprintf("--tag %s:%s", named.String(), "foo"))
		require.Contains(rst, fmt.Sprintf("--tag %s:%s", named.String(), "bar"))
	})

	t.Run("--build-arg", func(t *testing.T) {
		require := require.New(t)

		b, err := builder.NewDockerCmdBuilder(builder.BuilderConfig{
			DryRun: true,
		})
		require.NoError(err)

		o := bytes.Buffer{}
		err = b.Build(&clade.ResolvedImage{
			Named: named,
			Tags:  []string{"foo"},
			Args: map[string]string{
				"foo":    "bar",
				"answer": "42",
			},
			ContextPath: ".",
			Platform:    boolal.And("os", "arch"),
		}, builder.BuildOption{
			Stdout: &o,
			Stderr: io.Discard,
		})
		require.NoError(err)

		rst := o.String()
		require.Contains(rst, fmt.Sprintf("--build-arg %s=%s", "foo", "bar"))
		require.Contains(rst, fmt.Sprintf("--build-arg %s=%s", "answer", "42"))
	})
}

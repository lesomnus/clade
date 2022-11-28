package cmd_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/lesomnus/clade/cmd/clade/cmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootFlagsEvaluate(t *testing.T) {
	t.Run("valid log levels", func(t *testing.T) {
		require := require.New(t)

		levels := []string{
			"",
			"trace",
			"debug",
			"info",
			"warn",
			"error",
			"fatal",
		}

		flags := cmd.RootFlags{
			PortsPath: t.TempDir(),
		}

		for _, level := range levels {
			flags.LogLevel = level
			err := flags.Evaluate()
			require.NoError(err)
		}
	})

	t.Run("fails if", func(t *testing.T) {
		ports_dirpath := t.TempDir()
		temp_filepath := filepath.Join(ports_dirpath, "f")

		err := os.WriteFile(temp_filepath, []byte{}, 0644)
		require.NoError(t, err)

		tcx := []struct {
			desc  string
			flags cmd.RootFlags
			msgs  []string
		}{
			{
				desc: "invalid log level",
				flags: cmd.RootFlags{
					LogLevel:  "hard",
					PortsPath: ports_dirpath,
				},
				msgs: []string{"invalid", "hard"},
			},
			{
				desc: "ports path does not exists",
				flags: cmd.RootFlags{
					PortsPath: "not exists",
				},
				msgs: []string{"no such file or directory"},
			},
			{
				desc: "ports path not a directory",
				flags: cmd.RootFlags{
					PortsPath: temp_filepath,
				},
				msgs: []string{"not a directory"},
			},
		}
		for _, tc := range tcx {
			t.Run(tc.desc, func(t *testing.T) {
				require := require.New(t)

				err := tc.flags.Evaluate()
				require.Error(err)

				for _, msg := range tc.msgs {
					require.ErrorContains(err, msg)
				}
			})
		}
	})
}

func TestRootCmd(t *testing.T) {
	ports_dirpath := t.TempDir()

	t.Run("valid log levels", func(t *testing.T) {
		require := require.New(t)

		levels := []string{
			"trace",
			"debug",
			"info",
			"warn",
			"error",
			"fatal",
		}

		for _, level := range levels {
			next_cmd := cobra.Command{
				Use:  "mock",
				RunE: func(cmd *cobra.Command, args []string) error { return nil },
			}

			flags := cmd.RootFlags{}
			c := cmd.CreateRootCmd(&flags)
			c.AddCommand(&next_cmd)
			c.SetOut(io.Discard)
			c.SetArgs([]string{"--ports", ports_dirpath, "--log-level", level, "mock"})
			err := c.Execute()
			require.NoError(err)
			require.Equal(flags.LogLevel, level)
		}
	})

	t.Run("fails if", func(t *testing.T) {
		ports_dirpath := t.TempDir()

		tcs := []struct {
			desc string
			args []string
			msgs []string
		}{
			{
				desc: "ports directory does not exist",
				args: []string{"--ports", "not exist"},
				msgs: []string{"no such file or directory"},
			},
			{
				desc: "invalid log level",
				args: []string{"--ports", ports_dirpath, "--log-level", "easy"},
				msgs: []string{"invalid", "easy"},
			},
		}
		for _, tc := range tcs {
			t.Run(tc.desc, func(t *testing.T) {
				t.Run(tc.desc, func(t *testing.T) {
					require := require.New(t)

					next_cmd := cobra.Command{
						Use:  "mock",
						RunE: func(cmd *cobra.Command, args []string) error { return nil },
					}

					flags := cmd.RootFlags{}
					c := cmd.CreateRootCmd(&flags)
					c.AddCommand(&next_cmd)
					c.SetOut(io.Discard)
					c.SetArgs(append(tc.args, "mock"))
					err := c.Execute()
					require.Error(err)

					for _, msg := range tc.msgs {
						require.ErrorContains(err, msg)
					}
				})
			})
		}
	})
}

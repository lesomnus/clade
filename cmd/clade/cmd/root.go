package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/lesomnus/clade"
	"github.com/spf13/cobra"
)

type RootFlags struct {
	LogLevel  string
	PortsPath string
}

func (f *RootFlags) Evaluate() error {
	switch f.LogLevel {
	case "trace":
	case "debug":
	case "info":
	case "warn":
	case "error":
	case "fatal":
	case "":
		f.LogLevel = "warn"
	default:
		return fmt.Errorf("invalid log-level value: %s", f.LogLevel)
	}

	if p, err := clade.ResolvePath("", f.PortsPath, "ports"); err != nil {
		return fmt.Errorf("failed to resolve path to ports: %w", err)
	} else {
		f.PortsPath = p
	}

	if fi, err := os.Stat(f.PortsPath); err != nil {
		return err
	} else if !fi.IsDir() {
		return errors.New("path to ports not a directory: " + f.PortsPath)
	}

	return nil
}

func CreateCmdRoot(flags *RootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clade",
		Short: "Lade container images with your taste",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := root_flags.Evaluate(); err != nil {
				return err
			}

			// var level zerolog.Level
			// switch root_flags.logLevel {
			// case "trace":
			// 	level = zerolog.TraceLevel
			// case "debug":
			// 	level = zerolog.DebugLevel
			// case "info":
			// 	level = zerolog.InfoLevel
			// case "warn":
			// 	level = zerolog.WarnLevel
			// case "error":
			// 	level = zerolog.ErrorLevel
			// case "fatal":
			// 	level = zerolog.FatalLevel
			// default:
			// 	panic(fmt.Errorf("invalid log level: %s", root_flags.logLevel))
			// }

			// zerolog.Ctx(cmd.Context()).le

			// internal.Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(level)

			return nil
		},
	}

	cmd_flags := cmd.PersistentFlags()
	cmd_flags.StringVarP(&flags.LogLevel, "log-level", "l", "warn", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	cmd_flags.StringVar(&flags.PortsPath, "ports", "ports", "Path to repository")

	return cmd
}

var (
	root_flags RootFlags
	root_cmd   = CreateCmdRoot(&root_flags)
)

func Execute() error {
	return root_cmd.Execute()
}

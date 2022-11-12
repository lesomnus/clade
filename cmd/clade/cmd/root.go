package cmd

import (
	"fmt"
	"os"

	"github.com/lesomnus/clade/cmd/clade/cmd/internal"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var root_cmd = &cobra.Command{
	Use:   "clade",
	Short: "Lade container images with your taste",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := root_flags.Evaluate(); err != nil {
			return err
		}

		var level zerolog.Level
		switch root_flags.logLevel {
		case "trace":
			level = zerolog.TraceLevel
		case "debug":
			level = zerolog.DebugLevel
		case "info":
			level = zerolog.InfoLevel
		case "warn":
			level = zerolog.WarnLevel
		case "error":
			level = zerolog.ErrorLevel
		case "fatal":
			level = zerolog.FatalLevel
		default:
			panic(fmt.Errorf("invalid log level: %s", root_flags.logLevel))
		}

		internal.Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(level)

		return nil
	},
}

func Execute() error {
	return root_cmd.Execute()
}

func init() {
	flags := root_cmd.PersistentFlags()
	flags.StringVarP(&root_flags.logLevel, "log-level", "l", "warn", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	flags.StringVar(&root_flags.portsPath, "ports", "ports", "Path to repository")
}

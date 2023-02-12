package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/lesomnus/clade"
	"github.com/lesomnus/clade/cmd/clade/cmd/internal/client"
	"github.com/rs/zerolog"
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
		f.LogLevel = "info"
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

func CreateRootCmd(flags *RootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clade",
		Short: "Lade container images with your taste",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := flags.Evaluate(); err != nil {
				return err
			}

			var level zerolog.Level
			switch flags.LogLevel {
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
				panic(fmt.Errorf("invalid log level: %s", flags.LogLevel))
			}

			l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger().Level(level)
			cmd.SetContext(l.WithContext(cmd.Context()))

			docker_config_path := client.DefaultDockerConfigPath
			if _, err := os.Stat(docker_config_path); err == nil {
				l.Info().Str("path", docker_config_path).Msg("load credential from Docker config")
				auths, err := client.LoadAuthFromDockerConfig(docker_config_path)
				if err != nil {
					l.Warn().Msg("failed to load credential from Docker config")
				} else {
					for svc, auth := range auths {
						RegistryClient.Credentials.BasicAuths[svc] = auth
					}
				}
			}

			return nil
		},
	}

	cmd_flags := cmd.PersistentFlags()
	cmd_flags.StringVarP(&flags.LogLevel, "log-level", "l", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	cmd_flags.StringVar(&flags.PortsPath, "ports", "ports", "Path to repository")

	return cmd
}

var (
	root_flags RootFlags
	root_cmd   = CreateRootCmd(&root_flags)
)

func Execute() error {
	return root_cmd.Execute()
}

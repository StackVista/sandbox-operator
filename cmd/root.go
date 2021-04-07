package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hierynomus/taipan"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/internal/config"
)

func RootCommand() *cobra.Command {
	var verbosity int
	cmd := &cobra.Command{
		Use:   "sandboxer",
		Short: "StackState Sandbox operator",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			switch verbosity {
			case 0:
				// Nothing to do
			case 1:
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			case 2:
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			default:
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			}

			return nil
		},
	}

	cmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Print more verbose logging")

	return cmd
}

func Execute(ctx context.Context) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	cfg := &config.Config{}
	tp := taipan.New(&taipan.Config{
		DefaultConfigName:  "config",
		ConfigurationPaths: []string{".", "conf.d"},
		EnvironmentPrefix:  "SB",
		AddConfigFlag:      true,
		ConfigObject:       cfg,
	})

	cmd := RootCommand()
	cmd.AddCommand(SandboxCommand(cfg))
	cmd.AddCommand(ReaperCommand(cfg))
	cmd.AddCommand(ScalerCommand(cfg))
	tp.Inject(cmd)

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

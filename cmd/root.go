package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/hierynomus/taipan"
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/internal/config"
)

func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sandboxer",
		Short: "StackState Sandbox operator",
	}
}

func Execute(ctx context.Context) {
	config := &config.Config{}

	tp := taipan.New(&taipan.Config{
		DefaultConfigName:  "config",
		ConfigurationPaths: []string{".", "conf.d"},
		EnvironmentPrefix:  "SB",
		AddConfigFlag:      true,
		ConfigObject:       config,
	})

	cmd := RootCommand()
	cmd.AddCommand(SandboxCommand(config))
	cmd.AddCommand(ReaperCommand(config))
	cmd.AddCommand(ScalerCommand(config))

	tp.Inject(cmd)
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

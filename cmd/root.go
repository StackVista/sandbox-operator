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

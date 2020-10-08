package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func RootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sandboxer",
		Short: "StackState Sandbox operator",
	}
}

func Execute(ctx context.Context) {
	cmd := RootCommand()
	cmd.AddCommand(SandboxCommand())
	cmd.AddCommand(ReaperCommand())

	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

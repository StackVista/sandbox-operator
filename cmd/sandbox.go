package cmd

import (
	"github.com/stackvista/sandbox-operator/internal/sandbox"

	"github.com/spf13/cobra"
)

func SandboxCommand() *cobra.Command {
	config := &sandbox.OperatorConfig{}

	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Start the Sandbox controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sandbox.StartOperator(cmd.Context(), config)
		},
	}

	cmd.Flags().StringVarP(&config.MetricsAddr, "metrics-addr", "m", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVarP(&config.EnableLeaderElection, "enable-leader-election", "e", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	return cmd
}

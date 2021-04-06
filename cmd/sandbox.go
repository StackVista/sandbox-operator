package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/controllers/devops"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/pkg/operator"
)

func SandboxCommand(config *config.Config) *cobra.Command {
	operatorConfig := &operator.Config{}

	cmd := &cobra.Command{
		Use:   "sandbox",
		Short: "Start the Sandbox controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			return operator.StartOperator(cmd.Context(), operatorConfig, &devops.SandboxReconcilerFactory{
				Config: config,
			})
		},
	}

	cmd.Flags().StringVarP(&operatorConfig.MetricsAddr, "metrics-addr", "m", ":8080", "The address the metric endpoint binds to.")
	cmd.Flags().BoolVarP(&operatorConfig.EnableLeaderElection, "enable-leader-election", "e", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	return cmd
}

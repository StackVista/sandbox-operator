package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification/slack"
	"github.com/stackvista/sandbox-operator/internal/scaler"
)

func ScalerCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaler",
		Short: "Scaler scales down namespaces after a preconfigured timeout",
		RunE: func(cmd *cobra.Command, args []string) error {

			slack, err := slack.NewSlacker(cfg.Slack)
			if err != nil {
				return err
			}

			s, err := scaler.NewScaler(cmd.Context(), cfg.Scaler, slack)
			if err != nil {
				return err
			}

			return s.Run(cmd.Context())
		},
	}

	return cmd
}

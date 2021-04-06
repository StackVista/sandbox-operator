package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification/slack"
	"github.com/stackvista/sandbox-operator/internal/scaler"
)

func ScalerCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scaler",
		Short: "Scaler scales down namespaces after a preconfigured timeout",
		RunE: func(cmd *cobra.Command, args []string) error {
			slack, err := slack.NewSlacker(config.Slack)
			if err != nil {
				return err
			}

			s, err := scaler.NewScaler(cmd.Context(), config.Reaper, slack)
			if err != nil {
				return err
			}

			return s.Run(cmd.Context())
		},
	}

	return cmd
}

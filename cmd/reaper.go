package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/internal/config"
	"github.com/stackvista/sandbox-operator/internal/notification/slack"
	"github.com/stackvista/sandbox-operator/internal/reaper"
)

func ReaperCommand(config *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reaper",
		Short: "Reaper reaps namespaces that have exceeded their expiry date",
		RunE: func(cmd *cobra.Command, args []string) error {
			slack, err := slack.NewSlacker(config.Slack)
			if err != nil {
				return err
			}

			reaper, err := reaper.NewReaper(cmd.Context(), config.Reaper, slack)
			if err != nil {
				return err
			}

			return reaper.Run(cmd.Context())
		},
	}

	return cmd
}

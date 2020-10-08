package cmd

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"github.com/stackvista/sandbox-operator/internal/reaper"
)

func ReaperCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reaper",
		Short: "Reaper reaps namespaces that have exceeded their expiry date",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := &reaper.Config{}
			if err := envconfig.Process("", config); err != nil {
				return err
			}

			reaper, err := reaper.NewReaper(cmd.Context(), config)
			if err != nil {
				return err
			}

			reaper.Run(cmd.Context())
			return nil
		},
	}

	return cmd
}

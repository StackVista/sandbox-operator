package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(buildCerberusCommand())
}

func buildCerberusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reaper",
		Short: "Reaper reaps namespaces that have exceeded their expiry date",
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}

	return cmd
}

package org

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <org-name>",
		Short: "Get details of an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.Client()
			if err != nil {
				return err
			}
			p := cli.Printer()

			o, err := client.GetOrganization(args[0])
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(o.Name)
				return nil
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"name", o.Name},
				{"id", o.ID},
				{"execution_mode", o.ExecutionMode},
				{"agents_enabled", boolStr(o.AgentsEnabled)},
				{"drift_detection_enabled", boolStr(o.DriftDetectionEnabled)},
				{"is_trial", boolStr(o.IsTrial)},
				{"created_at", o.CreatedAt},
			}
			return p.Print(o, headers, rows)
		},
	}
}

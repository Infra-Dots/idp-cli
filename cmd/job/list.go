package job

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs for a workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := cli.Require("workspace", wsName); err != nil {
				return err
			}

			jobs, err := scope.Client.ListJobs(scope.Org, wsName)
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				for _, j := range jobs {
					scope.Printer.PrintID(j.ID)
				}
				return nil
			}

			headers := []string{"ID", "TYPE", "STATUS", "VIA", "UPDATED"}
			rows := make([][]string, len(jobs))
			for i, j := range jobs {
				rows[i] = []string{j.ID, j.Type, j.Status, j.Via, j.UpdatedDate}
			}
			return scope.Printer.Print(jobs, headers, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	return cmd
}

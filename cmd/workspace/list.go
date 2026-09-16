package workspace

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workspaces in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			workspaces, err := scope.Client.ListWorkspaces(scope.Org)
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				for _, ws := range workspaces {
					scope.Printer.PrintID(ws.Name)
				}
				return nil
			}

			headers := []string{"NAME", "SOURCE", "BRANCH", "TF VERSION", "AUTO APPLY", "AGENTS", "UPDATED"}
			rows := make([][]string, len(workspaces))
			for i, ws := range workspaces {
				rows[i] = []string{
					ws.Name,
					ws.Source,
					ws.Branch,
					ws.TerraformVersion,
					boolStr(ws.AutoApply),
					boolStr(ws.AgentsEnabled),
					ws.UpdatedDate,
				}
			}
			return scope.Printer.Print(workspaces, headers, rows)
		},
	}
}

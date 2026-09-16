package agent

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agent executions for an organization or workspace",
		Example: `  idp agent list --org my-org
  idp agent list --org my-org --workspace prod-infra`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			var history []api.AgentHistory
			if wsName != "" {
				history, err = scope.Client.ListWorkspaceAgentHistory(wsName)
			} else {
				history, err = scope.Client.ListAgentHistory(scope.Org)
			}
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				for _, h := range history {
					scope.Printer.PrintID(agentID(h))
				}
				return nil
			}

			rows := make([][]string, len(history))
			for i, h := range history {
				rows[i] = listRow(h)
			}
			return scope.Printer.Print(history, listHeaders, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Limit to one workspace")
	return cmd
}

package variable

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List variables (org-level, or workspace-level with --workspace)",
		Long: `List variables.

With --workspace, the listing shows the variables that workspace's runs actually
see: its own, plus the org-level variables it inherits. The SCOPE column tells
them apart.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			var vars []api.Variable
			if wsName != "" {
				vars, err = scope.Client.ListWorkspaceVariables(scope.Org, wsName)
			} else {
				vars, err = scope.Client.ListOrgVariables(scope.Org)
			}
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				for _, v := range vars {
					scope.Printer.PrintID(v.Key)
				}
				return nil
			}

			rows := make([][]string, len(vars))
			for i, v := range vars {
				rows[i] = listRow(v)
			}
			return scope.Printer.Print(vars, listHeaders, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name (omit for org-level variables)")
	return cmd
}

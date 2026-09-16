package variable

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <var-id>",
		Short: "Delete a variable by ID",
		Long: `Delete a variable by ID.

Variable IDs are unique across the organization, so this works for both
org-level and workspace-level variables. Use ` + "`idp variable list`" + ` to find an ID.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := scope.Client.DeleteVariable(scope.Org, args[0]); err != nil {
				return err
			}
			fmt.Printf("Variable %s deleted\n", args[0])
			return nil
		},
	}
	return cmd
}

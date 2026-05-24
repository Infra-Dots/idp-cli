package variable

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newDeleteCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "delete <var-id>",
		Short: "Delete a variable by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))

			var err error
			if wsName != "" {
				err = client.DeleteWorkspaceVariable(orgName, wsName, args[0])
			} else {
				err = client.DeleteOrgVariable(orgName, args[0])
			}
			if err != nil {
				return err
			}
			fmt.Printf("Variable %s deleted\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name (omit for org-level)")
	return cmd
}

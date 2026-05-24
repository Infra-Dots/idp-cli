package variable

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newListCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List variables (org-level, or workspace-level with --workspace)",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			var vars []api.Variable
			var err error
			if wsName != "" {
				vars, err = client.ListWorkspaceVariables(orgName, wsName)
			} else {
				vars, err = client.ListOrgVariables(orgName)
			}
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, v := range vars {
					p.PrintID(v.Key)
				}
				return nil
			}

			headers := []string{"KEY", "VALUE", "SENSITIVE", "HCL", "ID"}
			rows := make([][]string, len(vars))
			for i, v := range vars {
				val := v.Value
				if v.Sensitive {
					val = "***"
				}
				rows[i] = []string{v.Key, val, boolStr(v.Sensitive), boolStr(v.HCL), v.ID}
			}
			return p.Print(vars, headers, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name (omit for org-level variables)")
	return cmd
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

package variable

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newSetCmd() *cobra.Command {
	var wsName string
	var sensitive, hcl bool

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Create or update a variable",
		Example: `  idp variable set AWS_REGION us-east-1 --org my-org
  idp variable set TF_VAR_db_password secret --org my-org --workspace prod --sensitive`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			in := api.SetVariableInput{
				Key:       args[0],
				Value:     args[1],
				Sensitive: sensitive,
				HCL:       hcl,
			}

			var v *api.Variable
			var err error
			if wsName != "" {
				v, err = client.CreateWorkspaceVariable(orgName, wsName, in)
			} else {
				v, err = client.CreateOrgVariable(orgName, in)
			}
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(v.ID)
				return nil
			}

			headers := []string{"KEY", "VALUE", "SENSITIVE", "HCL", "ID"}
			val := v.Value
			if v.Sensitive {
				val = "***"
			}
			rows := [][]string{{v.Key, val, boolStr(v.Sensitive), boolStr(v.HCL), v.ID}}
			return p.Print(v, headers, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name (omit for org-level)")
	cmd.Flags().BoolVar(&sensitive, "sensitive", false, "Mark value as sensitive (write-only)")
	cmd.Flags().BoolVar(&hcl, "hcl", false, "Parse value as HCL")
	return cmd
}

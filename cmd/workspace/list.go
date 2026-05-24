package workspace

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List workspaces in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			workspaces, err := client.ListWorkspaces(orgName)
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, ws := range workspaces {
					p.PrintID(ws.Name)
				}
				return nil
			}

			headers := []string{"NAME", "TF VERSION", "AUTO APPLY", "AGENTS", "UPDATED"}
			rows := make([][]string, len(workspaces))
			for i, ws := range workspaces {
				rows[i] = []string{
					ws.Name,
					ws.TerraformVersion,
					boolStr(ws.AutoApply),
					boolStr(ws.AgentsEnabled),
					ws.UpdatedAt,
				}
			}
			return p.Print(workspaces, headers, rows)
		},
	}
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

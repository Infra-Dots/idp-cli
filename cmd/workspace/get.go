package workspace

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <workspace>",
		Short: "Get details of a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			ws, err := client.GetWorkspace(orgName, args[0])
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(ws.Name)
				return nil
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"name", ws.Name},
				{"id", ws.ID},
				{"terraform_version", ws.TerraformVersion},
				{"auto_apply", boolStr(ws.AutoApply)},
				{"agents_enabled", boolStr(ws.AgentsEnabled)},
				{"created_at", ws.CreatedAt},
				{"updated_at", ws.UpdatedAt},
			}
			return p.Print(ws, headers, rows)
		},
	}
}

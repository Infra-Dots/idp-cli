package workspace

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newUpdateCmd() *cobra.Command {
	var tfVersion string
	var autoApply, agentsEnabled bool

	cmd := &cobra.Command{
		Use:   "update <workspace>",
		Short: "Update workspace settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			in := api.UpdateWorkspaceInput{
				TerraformVersion: tfVersion,
			}
			if cmd.Flags().Changed("auto-apply") {
				in.AutoApply = &autoApply
			}
			if cmd.Flags().Changed("agents-enabled") {
				in.AgentsEnabled = &agentsEnabled
			}

			ws, err := client.UpdateWorkspace(orgName, args[0], in)
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
				{"updated_at", ws.UpdatedAt},
			}
			return p.Print(ws, headers, rows)
		},
	}

	cmd.Flags().StringVar(&tfVersion, "tf-version", "", "Terraform/OpenTofu version")
	cmd.Flags().BoolVar(&autoApply, "auto-apply", false, "Enable/disable auto-apply")
	cmd.Flags().BoolVar(&agentsEnabled, "agents-enabled", false, "Enable/disable AI agents")

	return cmd
}

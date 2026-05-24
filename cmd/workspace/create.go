package workspace

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newCreateCmd() *cobra.Command {
	var (
		name             string
		vcsID            string
		repo             string
		tfVersion        string
		autoApply        bool
		agentsEnabled    bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		Example: `  idp workspace create --org my-org --name prod-infra --vcs abc123 --repo my-org/infra
  idp workspace create --org my-org --name dev --tf-version 1.9.0 --auto-apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			in := api.CreateWorkspaceInput{
				Name:             name,
				VcsID:            vcsID,
				Repository:       repo,
				TerraformVersion: tfVersion,
			}
			if cmd.Flags().Changed("auto-apply") {
				in.AutoApply = &autoApply
			}
			if cmd.Flags().Changed("agents-enabled") {
				in.AgentsEnabled = &agentsEnabled
			}

			ws, err := client.CreateWorkspace(orgName, in)
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
			}
			return p.Print(ws, headers, rows)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Workspace name")
	cmd.Flags().StringVar(&vcsID, "vcs", "", "VCS connection ID")
	cmd.Flags().StringVar(&repo, "repo", "", "Repository path (e.g. org/repo)")
	cmd.Flags().StringVar(&tfVersion, "tf-version", "", "Terraform/OpenTofu version")
	cmd.Flags().BoolVar(&autoApply, "auto-apply", false, "Automatically apply after a successful plan")
	cmd.Flags().BoolVar(&agentsEnabled, "agents-enabled", false, "Enable AI agents for this workspace")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

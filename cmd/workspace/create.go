package workspace

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
)

func newCreateCmd() *cobra.Command {
	var (
		name          string
		vcsID         string
		source        string
		branch        string
		folder        string
		description   string
		tfVersion     string
		iacType       string
		autoApply     bool
		agentsEnabled bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new workspace",
		Long: `Create a new workspace.

--source, --branch and --tf-version are required by the API. --source is the
repository the workspace tracks, in the form the VCS provider expects
(e.g. my-org/infra).`,
		Example: `  idp workspace create --org my-org --name prod-infra \
    --vcs abc123 --source my-org/infra --branch main --tf-version 1.9.0
  idp workspace create --org my-org --name dev \
    --source my-org/infra --branch dev --tf-version 1.9.0 --folder /envs/dev --auto-apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := cli.Require("tf-version", tfVersion); err != nil {
				return err
			}

			in := api.CreateWorkspaceInput{
				Name:             name,
				Source:           source,
				Branch:           branch,
				Folder:           folder,
				Description:      description,
				TerraformVersion: tfVersion,
				IacType:          iacType,
				VcsID:            vcsID,
			}
			if cmd.Flags().Changed("auto-apply") {
				in.AutoApply = &autoApply
			}
			if cmd.Flags().Changed("agents-enabled") {
				in.AgentsEnabled = &agentsEnabled
			}

			ws, err := scope.Client.CreateWorkspace(scope.Org, in)
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				scope.Printer.PrintID(ws.Name)
				return nil
			}
			return scope.Printer.Print(ws, detailHeaders, detailRows(ws))
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Workspace name")
	cmd.Flags().StringVar(&vcsID, "vcs", "", "VCS connection ID")
	cmd.Flags().StringVar(&source, "source", "", "Repository the workspace tracks (e.g. org/repo)")
	cmd.Flags().StringVar(&branch, "branch", "main", "Branch to track")
	cmd.Flags().StringVar(&folder, "folder", "/", "Folder within the repository")
	cmd.Flags().StringVar(&description, "description", "", "Workspace description")
	cmd.Flags().StringVar(&tfVersion, "tf-version", "", "Terraform/OpenTofu version (required)")
	cmd.Flags().StringVar(&iacType, "iac-type", "", "IaC tool: terraform, opentofu")
	cmd.Flags().BoolVar(&autoApply, "auto-apply", false, "Automatically apply after a successful plan")
	cmd.Flags().BoolVar(&agentsEnabled, "agents-enabled", false, "Enable AI agents for this workspace")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("source")

	return cmd
}

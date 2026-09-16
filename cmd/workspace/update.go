package workspace

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
)

func newUpdateCmd() *cobra.Command {
	var tfVersion, source, branch, folder, description string
	var autoApply, agentsEnabled bool

	cmd := &cobra.Command{
		Use:   "update <workspace>",
		Short: "Update workspace settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			in := api.UpdateWorkspaceInput{
				TerraformVersion: tfVersion,
				Source:           source,
				Branch:           branch,
				Folder:           folder,
				Description:      description,
			}
			if cmd.Flags().Changed("auto-apply") {
				in.AutoApply = &autoApply
			}
			if cmd.Flags().Changed("agents-enabled") {
				in.AgentsEnabled = &agentsEnabled
			}

			ws, err := scope.Client.UpdateWorkspace(scope.Org, args[0], in)
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

	cmd.Flags().StringVar(&tfVersion, "tf-version", "", "Terraform/OpenTofu version")
	cmd.Flags().StringVar(&source, "source", "", "Repository the workspace tracks (e.g. org/repo)")
	cmd.Flags().StringVar(&branch, "branch", "", "Branch to track")
	cmd.Flags().StringVar(&folder, "folder", "", "Folder within the repository")
	cmd.Flags().StringVar(&description, "description", "", "Workspace description")
	cmd.Flags().BoolVar(&autoApply, "auto-apply", false, "Enable/disable auto-apply")
	cmd.Flags().BoolVar(&agentsEnabled, "agents-enabled", false, "Enable/disable AI agents")

	return cmd
}

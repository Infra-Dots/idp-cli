package workspace

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
)

// NewCmd returns the `idp workspace` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Manage workspaces",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newUpdateCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

// detailHeaders and detailRows render a single workspace the same way for
// create, get and update.
var detailHeaders = []string{"FIELD", "VALUE"}

func detailRows(ws *api.Workspace) [][]string {
	return [][]string{
		{"name", ws.Name},
		{"id", ws.ID},
		{"source", ws.Source},
		{"branch", ws.Branch},
		{"folder", ws.Folder},
		{"terraform_version", ws.TerraformVersion},
		{"iac_type", ws.IacType},
		{"execution_mode", ws.ExecutionMode},
		{"auto_apply", boolStr(ws.AutoApply)},
		{"agents_enabled", boolStr(ws.AgentsEnabled)},
		{"locked", boolStr(ws.Locked)},
		{"created_date", ws.CreatedDate},
		{"updated_date", ws.UpdatedDate},
	}
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

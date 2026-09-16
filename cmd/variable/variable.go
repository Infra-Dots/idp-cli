package variable

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
)

// NewCmd returns the `idp variable` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variable",
		Aliases: []string{"var"},
		Short:   "Manage org and workspace variables",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newSetCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

var listHeaders = []string{"KEY", "VALUE", "CATEGORY", "SCOPE", "SENSITIVE", "HCL", "ID"}

func listRow(v api.Variable) []string {
	scope := "org"
	if v.Workspace != "" {
		scope = "workspace"
	}
	// The API already blanks sensitive values; label them so an empty cell
	// doesn't read as an unset variable.
	value := v.Value
	if v.Sensitive {
		value = "***"
	}
	return []string{v.Key, value, v.Category, scope, boolStr(v.Sensitive), boolStr(v.HCL), v.ID}
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

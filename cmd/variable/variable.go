package variable

import "github.com/spf13/cobra"

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

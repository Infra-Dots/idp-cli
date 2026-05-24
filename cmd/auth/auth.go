package auth

import "github.com/spf13/cobra"

// NewCmd returns the `idp auth` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate and manage API tokens",
	}
	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newTokenCmd())
	return cmd
}

package vcs

import "github.com/spf13/cobra"

// NewCmd returns the `idp vcs` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vcs",
		Short: "Manage VCS connections",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newCreateCmd())
	cmd.AddCommand(newDeleteCmd())
	return cmd
}

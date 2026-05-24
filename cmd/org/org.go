package org

import "github.com/spf13/cobra"

// NewCmd returns the `idp org` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage organizations",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newGetCmd())
	return cmd
}

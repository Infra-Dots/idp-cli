package agent

import "github.com/spf13/cobra"

// NewCmd returns the `idp agent` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "View AI agent execution history",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newHistoryCmd())
	return cmd
}

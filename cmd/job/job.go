package job

import "github.com/spf13/cobra"

// NewCmd returns the `idp job` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage workspace jobs (plan/apply)",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newApproveCmd())
	cmd.AddCommand(newCancelCmd())
	cmd.AddCommand(newDiscardCmd())
	cmd.AddCommand(newOutputCmd())
	return cmd
}

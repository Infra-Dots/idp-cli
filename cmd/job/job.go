package job

import "github.com/spf13/cobra"

// NewCmd returns the `idp job` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage workspace jobs (plan/apply)",
	}

	// One --workspace for every subcommand, so a command line that works for
	// `get` works for `output` too. list, run, get and cancel need it; approve,
	// discard and output reach the job by its ID alone and don't use it.
	var wsName string
	cmd.PersistentFlags().StringVarP(&wsName, "workspace", "w", "",
		"Workspace name (required by list, run, get and cancel)")

	cmd.AddCommand(newListCmd(&wsName))
	cmd.AddCommand(newRunCmd(&wsName))
	cmd.AddCommand(newGetCmd(&wsName))
	cmd.AddCommand(newApproveCmd())
	cmd.AddCommand(newCancelCmd(&wsName))
	cmd.AddCommand(newDiscardCmd())
	cmd.AddCommand(newOutputCmd())
	return cmd
}

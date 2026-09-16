package workspace

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <workspace>",
		Short: "Delete a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			org, err := cli.RequireOrg()
			if err != nil {
				return err
			}

			if !force {
				fmt.Printf("Delete workspace %q in org %q? This cannot be undone. [y/N]: ", args[0], org)
				var confirm string
				_, _ = fmt.Fscan(os.Stdin, &confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			client, err := cli.Client()
			if err != nil {
				return err
			}
			if err := client.DeleteWorkspace(org, args[0]); err != nil {
				return err
			}
			fmt.Printf("Workspace %q deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

package workspace

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <workspace>",
		Short: "Delete a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}

			if !force {
				fmt.Printf("Delete workspace %q in org %q? This cannot be undone. [y/N]: ", args[0], orgName)
				var confirm string
				_, _ = fmt.Fscan(os.Stdin, &confirm)
				if confirm != "y" && confirm != "Y" {
					fmt.Println("Aborted.")
					return nil
				}
			}

			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			if err := client.DeleteWorkspace(orgName, args[0]); err != nil {
				return err
			}
			fmt.Printf("Workspace %q deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}

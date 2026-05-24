package vcs

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <vcs-id>",
		Short: "Delete a VCS connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			if err := client.DeleteVCS(orgName, args[0]); err != nil {
				return err
			}
			fmt.Printf("VCS connection %s deleted\n", args[0])
			return nil
		},
	}
}

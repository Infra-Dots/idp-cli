package vcs

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <vcs-id>",
		Short: "Delete a VCS connection",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := scope.Client.DeleteVCS(scope.Org, args[0]); err != nil {
				return err
			}
			fmt.Printf("VCS connection %s deleted\n", args[0])
			return nil
		},
	}
}

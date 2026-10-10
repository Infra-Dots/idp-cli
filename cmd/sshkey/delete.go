package sshkey

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <name|id>",
		Short: "Delete an SSH key (workspaces using it keep running, without it)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			key, err := scope.Client.ResolveSSHKey(scope.Org, args[0])
			if err != nil {
				return err
			}
			if err := scope.Client.DeleteSSHKey(scope.Org, key.ID); err != nil {
				return err
			}
			fmt.Printf("SSH key %s (%s) deleted\n", key.Name, key.Fingerprint)
			return nil
		},
	}
}

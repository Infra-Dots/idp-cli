package sshkey

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SSH keys (never the private keys)",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			keys, err := scope.Client.ListSSHKeys(scope.Org)
			if err != nil {
				return err
			}
			if scope.Printer.Quiet {
				for _, k := range keys {
					scope.Printer.PrintID(k.ID)
				}
				return nil
			}
			rows := make([][]string, len(keys))
			for i, k := range keys {
				rows[i] = listRow(k)
			}
			return scope.Printer.Print(keys, listHeaders, rows)
		},
	}
}

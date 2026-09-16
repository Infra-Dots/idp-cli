package vcs

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List VCS connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			connections, err := scope.Client.ListVCS(scope.Org)
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				for _, v := range connections {
					scope.Printer.PrintID(v.ID)
				}
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "STATUS", "CREATED"}
			rows := make([][]string, len(connections))
			for i, v := range connections {
				rows[i] = []string{v.ID, v.Name, v.VCSType, v.Status, v.CreatedDate}
			}
			return scope.Printer.Print(connections, headers, rows)
		},
	}
}

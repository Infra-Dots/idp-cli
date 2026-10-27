package analyze

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List repository analyses for an organization",
		Example: `  idp analyze list --org my-org`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			items, raws, err := scope.Client.ListAnalyses(scope.Org)
			if err != nil {
				return err
			}
			if scope.Printer.Quiet {
				for _, s := range items {
					scope.Printer.PrintID(s.ID)
				}
				return nil
			}
			rows := make([][]string, len(items))
			for i, s := range items {
				rows[i] = listRow(s)
			}
			return scope.Printer.Print(rawsOrItems(scope, raws, items), listHeaders, rows)
		},
	}
}

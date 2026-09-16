package workspace

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <workspace>",
		Short: "Get details of a workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			ws, err := scope.Client.GetWorkspace(scope.Org, args[0])
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				scope.Printer.PrintID(ws.Name)
				return nil
			}
			return scope.Printer.Print(ws, detailHeaders, detailRows(ws))
		},
	}
}

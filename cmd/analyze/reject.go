package analyze

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

func newRejectCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "reject ID",
		Short:   "Reject a proposal; nothing is requested",
		Example: `  idp analyze reject 6f1c0c1e-...`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			a, err := scope.Client.RejectAnalysis(scope.Org, args[0])
			if err != nil {
				return err
			}
			if scope.Printer.Format != output.FormatTable {
				return printRaw(scope.Printer, a.Raw, a)
			}
			_, _ = fmt.Fprintf(scope.Printer.Out, "Analysis %s is now %s.\n", a.ID, a.Status)
			return nil
		},
	}
}

package analyze

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "show ID",
		Short:   "Show an analysis: status, decisions, questions and the targets it would request",
		Example: `  idp analyze show 6f1c0c1e-...`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			a, err := scope.Client.GetAnalysis(scope.Org, args[0])
			if err != nil {
				return err
			}
			return showAnalysis(scope, a)
		},
	}
}

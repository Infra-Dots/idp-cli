package agent

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history <job-id>",
		Short: "Show the agent executions for one job",
		Long: `Show the agent executions for one job.

A job can have more than one — typically a review, then an implementation — so
this always prints a list.`,
		Example: `  idp agent history 6f1c2b9e-... --org my-org`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			history, err := scope.Client.GetJobAgentHistory(args[0])
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				for _, h := range history {
					scope.Printer.PrintID(agentID(h))
				}
				return nil
			}

			rows := make([][]string, len(history))
			for i, h := range history {
				rows[i] = listRow(h)
			}
			return scope.Printer.Print(history, listHeaders, rows)
		},
	}
}

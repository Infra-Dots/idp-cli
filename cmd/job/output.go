package job

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newOutputCmd() *cobra.Command {
	var stage string

	cmd := &cobra.Command{
		Use:   "output <job-id>",
		Short: "Print the output log of a job",
		Long: `Print the output log of a job.

With no --stage, prints every stage of the run in execution order.`,
		Example: `  idp job output <job-id> --org my-org
  idp job output <job-id> --org my-org --stage apply`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// The stage endpoints are scoped by org membership server-side and
			// keyed on the job ID alone, so no --workspace is needed.
			client, err := cli.Client()
			if err != nil {
				return err
			}

			if stage != "" {
				js, err := client.GetJobStage(args[0], stage)
				if err != nil {
					return err
				}
				fmt.Print(js.Output)
				return nil
			}

			stages, err := client.ListJobStages(args[0])
			if err != nil {
				return err
			}
			if len(stages) == 0 {
				fmt.Fprintln(os.Stderr, "job has no stages yet")
				return nil
			}
			for _, s := range stages {
				// Headers go to stderr so `idp job output <id> > run.log` still
				// captures exactly the log text.
				fmt.Fprintf(os.Stderr, "\n=== %s ===\n", strings.ToUpper(s.Type))
				fmt.Print(s.Output)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&stage, "stage", "", "Stage to fetch: init, validate, tflint, plan, apply (default: all)")
	return cmd
}

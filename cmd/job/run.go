package job

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

func newRunCmd() *cobra.Command {
	var wsName, jobType string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Trigger a new plan or apply job",
		Example: `  idp job run --org my-org --workspace prod-infra
  idp job run --org my-org --workspace prod-infra --type apply
  idp job run --org my-org --workspace prod-infra --type destroy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := cli.Require("workspace", wsName); err != nil {
				return err
			}
			// Catch a bad type here rather than as an opaque 400 from the API.
			if !api.ValidJobType(jobType) {
				return output.NewError("invalid --type %q: must be one of %s",
					jobType, strings.Join(api.JobTypes, ", "))
			}

			j, err := scope.Client.CreateJob(scope.Org, wsName, api.CreateJobInput{Type: jobType})
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				scope.Printer.PrintID(j.ID)
				return nil
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"id", j.ID},
				{"type", j.Type},
				{"status", j.Status},
				{"created_date", j.CreatedDate},
			}
			return scope.Printer.Print(j, headers, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	cmd.Flags().StringVar(&jobType, "type", "plan", "Job type: "+strings.Join(api.JobTypes, ", "))
	return cmd
}

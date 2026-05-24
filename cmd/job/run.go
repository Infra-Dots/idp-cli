package job

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newRunCmd() *cobra.Command {
	var wsName, jobType string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Trigger a new plan or apply job",
		Example: `  idp job run --org my-org --workspace prod-infra
  idp job run --org my-org --workspace prod-infra --type apply`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			if wsName == "" {
				return output.NewError("--workspace is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			j, err := client.CreateJob(orgName, wsName, api.CreateJobInput{Type: jobType})
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(j.ID)
				return nil
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"id", j.ID},
				{"type", j.Type},
				{"status", j.Status},
				{"created_at", j.CreatedAt},
			}
			return p.Print(j, headers, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	cmd.Flags().StringVar(&jobType, "type", "plan", "Job type: plan, apply, plan_only")
	return cmd
}

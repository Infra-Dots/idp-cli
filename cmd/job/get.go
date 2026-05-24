package job

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

// terminalStatuses mirrors idp.workers.models.JobStatuses terminal values.
var terminalStatuses = map[string]bool{
	"completed": true,
	"applied":   true,
	"rejected":  true,
	"failed":    true,
	"cancelled": true,
}

// failedStatuses cause a non-zero exit when using --watch.
var failedStatuses = map[string]bool{
	"rejected":  true,
	"failed":    true,
	"cancelled": true,
}

func newGetCmd() *cobra.Command {
	var wsName string
	var watch bool
	var interval int

	cmd := &cobra.Command{
		Use:   "get <job-id>",
		Short: "Get details of a job",
		Example: `  idp job get <job-id> --org my-org --workspace prod-infra
  idp job get <job-id> --org my-org --workspace prod-infra --watch
  idp job get <job-id> --org my-org --workspace prod-infra --watch --interval 5`,
		Args: cobra.ExactArgs(1),
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

			if !watch {
				return getAndPrint(client, p, orgName, wsName, args[0])
			}

			// Watch mode: poll until terminal state.
			return watchJob(client, orgName, wsName, args[0], interval)
		},
	}

	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until the job reaches a terminal state")
	cmd.Flags().IntVar(&interval, "interval", 3, "Polling interval in seconds (used with --watch)")
	return cmd
}

func getAndPrint(client *api.Client, p *output.Printer, orgName, wsName, jobID string) error {
	j, err := client.GetJob(orgName, wsName, jobID)
	if err != nil {
		return err
	}

	if p.Quiet {
		p.PrintID(j.Status)
		return nil
	}

	headers := []string{"FIELD", "VALUE"}
	rows := [][]string{
		{"id", j.ID},
		{"type", j.Type},
		{"status", j.Status},
		{"created_by", j.CreatedBy},
		{"created_at", j.CreatedAt},
		{"updated_at", j.UpdatedAt},
	}
	return p.Print(j, headers, rows)
}

func watchJob(client *api.Client, orgName, wsName, jobID string, intervalSecs int) error {
	ticker := time.NewTicker(time.Duration(intervalSecs) * time.Second)
	defer ticker.Stop()

	var lastStatus string

	// Print immediately, then on each tick.
	poll := func() (*api.Job, error) {
		return client.GetJob(orgName, wsName, jobID)
	}

	for {
		j, err := poll()
		if err != nil {
			return err
		}

		if j.Status != lastStatus {
			lastStatus = j.Status
			fmt.Fprintf(os.Stdout, "[%s] job %s  status: %s\n",
				time.Now().Format("15:04:05"), j.ID, j.Status)
		}

		if terminalStatuses[j.Status] {
			if failedStatuses[j.Status] {
				fmt.Fprintf(os.Stderr, "job finished with status: %s\n", j.Status)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "job finished with status: %s\n", j.Status)
			return nil
		}

		<-ticker.C
	}
}

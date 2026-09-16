package job

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

// Exit codes for --watch, so CI can tell the three outcomes apart.
const (
	exitJobFailed        = 1
	exitJobNeedsApproval = 2
	exitWatchTimeout     = 3
)

// Terminal states, mirroring idp.workers.models.JobStatuses. Note the literal
// spelling of "Waiting Approval" — it has a space and capitals, unlike every
// other status.
const statusWaitingApproval = "Waiting Approval"

// terminalStatuses are the states a job never leaves on its own. "Waiting
// Approval" belongs here: without it, watching a plan that needs approval polls
// forever, because approval only ever arrives from outside.
var terminalStatuses = map[string]bool{
	"completed":           true,
	"applied":             true,
	"rejected":            true,
	"failed":              true,
	"cancelled":           true,
	statusWaitingApproval: true,
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
	var interval, timeout int

	cmd := &cobra.Command{
		Use:   "get <job-id>",
		Short: "Get details of a job",
		Long: `Get details of a job.

With --watch, poll until the job reaches a state it will not leave on its own,
then exit: 0 if it succeeded, 1 if it failed or was cancelled, 2 if it is
waiting for approval, 3 if --timeout elapsed first.`,
		Example: `  idp job get <job-id> --org my-org --workspace prod-infra
  idp job get <job-id> --org my-org --workspace prod-infra --watch
  idp job get <job-id> --org my-org --workspace prod-infra --watch --interval 5 --timeout 1800`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := cli.Require("workspace", wsName); err != nil {
				return err
			}

			if !watch {
				return getAndPrint(scope, wsName, args[0])
			}
			return watchJob(scope.Client, scope.Org, wsName, args[0], interval, timeout)
		},
	}

	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll until the job reaches a terminal state")
	cmd.Flags().IntVar(&interval, "interval", 3, "Polling interval in seconds (used with --watch)")
	cmd.Flags().IntVar(&timeout, "timeout", 3600, "Give up after this many seconds (used with --watch; 0 waits forever)")
	return cmd
}

func getAndPrint(scope *cli.OrgScope, wsName, jobID string) error {
	j, err := scope.Client.GetJob(scope.Org, wsName, jobID)
	if err != nil {
		return err
	}

	if scope.Printer.Quiet {
		scope.Printer.PrintID(j.Status)
		return nil
	}

	headers := []string{"FIELD", "VALUE"}
	rows := [][]string{
		{"id", j.ID},
		{"type", j.Type},
		{"status", j.Status},
		{"created_by", j.CreatedBy},
		{"via", j.Via},
		{"created_date", j.CreatedDate},
		{"updated_date", j.UpdatedDate},
	}
	return scope.Printer.Print(j, headers, rows)
}

func watchJob(client *api.Client, orgName, wsName, jobID string, intervalSecs, timeoutSecs int) error {
	ticker := time.NewTicker(time.Duration(intervalSecs) * time.Second)
	defer ticker.Stop()

	// A zero timeout means wait indefinitely; a nil channel never fires.
	var deadline <-chan time.Time
	if timeoutSecs > 0 {
		t := time.NewTimer(time.Duration(timeoutSecs) * time.Second)
		defer t.Stop()
		deadline = t.C
	}

	var lastStatus string

	for {
		j, err := client.GetJob(orgName, wsName, jobID)
		if err != nil {
			return err
		}

		if j.Status != lastStatus {
			lastStatus = j.Status
			fmt.Fprintf(os.Stdout, "[%s] job %s  status: %s\n",
				time.Now().Format("15:04:05"), j.ID, j.Status)
		}

		if terminalStatuses[j.Status] {
			switch {
			case failedStatuses[j.Status]:
				output.Err("job finished with status: %s", j.Status)
				os.Exit(exitJobFailed)
			case j.Status == statusWaitingApproval:
				// Not a failure — the job is fine and needs a human. Distinct
				// exit code so a pipeline can gate on it rather than fail.
				fmt.Fprintf(os.Stdout, "job is waiting for approval — run `idp job approve %s`\n", j.ID)
				os.Exit(exitJobNeedsApproval)
			}
			fmt.Fprintf(os.Stdout, "job finished with status: %s\n", j.Status)
			return nil
		}

		select {
		case <-ticker.C:
		case <-deadline:
			output.Err("timed out after %ds waiting for job %s (last status: %s)",
				timeoutSecs, jobID, lastStatus)
			os.Exit(exitWatchTimeout)
		}
	}
}

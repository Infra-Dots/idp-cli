package agent

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
)

// NewCmd returns the `idp agent` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Run the InfraDots agent locally, and view agent history",
	}
	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newHistoryCmd())
	cmd.AddCommand(newReviewCmd())
	return cmd
}

var listHeaders = []string{"ID", "TYPE", "STATUS", "IMPLEMENTED", "PR", "TIMESTAMP"}

func listRow(h api.AgentHistory) []string {
	return []string{agentID(h), h.Type, h.Status, boolStr(h.Implemented), h.PRURL, h.Timestamp}
}

// agentID renders the run's integer id for display and for --quiet piping.
func agentID(h api.AgentHistory) string {
	return strconv.Itoa(h.ID)
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

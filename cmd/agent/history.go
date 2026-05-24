package agent

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newHistoryCmd() *cobra.Command {
	var jobID string

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Get agent execution history for a job",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			// If a job ID is given fetch it directly, otherwise list org history.
			if jobID != "" {
				ah, err := client.GetAgentHistory(jobID)
				if err != nil {
					return err
				}
				if p.Quiet {
					p.PrintID(ah.ID)
					return nil
				}
				headers := []string{"FIELD", "VALUE"}
				rows := [][]string{
					{"id", ah.ID},
					{"type", ah.Type},
					{"status", ah.Status},
					{"created_at", ah.CreatedAt},
					{"updated_at", ah.UpdatedAt},
				}
				return p.Print(ah, headers, rows)
			}

			history, err := client.ListAgentHistory(orgName)
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, h := range history {
					p.PrintID(h.ID)
				}
				return nil
			}

			headers := []string{"ID", "TYPE", "STATUS", "UPDATED"}
			rows := make([][]string, len(history))
			for i, h := range history {
				rows[i] = []string{h.ID, h.Type, h.Status, h.UpdatedAt}
			}
			return p.Print(history, headers, rows)
		},
	}
	cmd.Flags().StringVar(&jobID, "job", "", "Fetch history for a specific job ID")
	return cmd
}

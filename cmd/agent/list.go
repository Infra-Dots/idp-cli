package agent

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List agent executions for an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

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
}

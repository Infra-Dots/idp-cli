package job

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newListCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs for a workspace",
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

			jobs, err := client.ListJobs(orgName, wsName)
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, j := range jobs {
					p.PrintID(j.ID)
				}
				return nil
			}

			headers := []string{"ID", "TYPE", "STATUS", "CREATED BY", "UPDATED"}
			rows := make([][]string, len(jobs))
			for i, j := range jobs {
				rows[i] = []string{j.ID, j.Type, j.Status, j.CreatedBy, j.UpdatedAt}
			}
			return p.Print(jobs, headers, rows)
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	return cmd
}

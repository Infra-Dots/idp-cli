package vcs

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List VCS connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			connections, err := client.ListVCS(orgName)
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, v := range connections {
					p.PrintID(v.ID)
				}
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "CREATED"}
			rows := make([][]string, len(connections))
			for i, v := range connections {
				rows[i] = []string{v.ID, v.Name, v.VCSType, v.CreatedAt}
			}
			return p.Print(connections, headers, rows)
		},
	}
}

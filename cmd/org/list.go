package org

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			orgs, err := client.ListOrganizations()
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, o := range orgs {
					p.PrintID(o.Name)
				}
				return nil
			}

			headers := []string{"NAME", "DISPLAY NAME", "ID", "CREATED"}
			rows := make([][]string, len(orgs))
			for i, o := range orgs {
				rows[i] = []string{o.Name, o.DisplayName, o.ID, o.CreatedAt}
			}
			return p.Print(orgs, headers, rows)
		},
	}
}

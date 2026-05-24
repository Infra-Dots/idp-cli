package org

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <org-name>",
		Short: "Get details of an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			o, err := client.GetOrganization(args[0])
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(o.Name)
				return nil
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"name", o.Name},
				{"display_name", o.DisplayName},
				{"id", o.ID},
				{"created_at", o.CreatedAt},
			}
			return p.Print(o, headers, rows)
		},
	}
}

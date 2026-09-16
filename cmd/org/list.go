package org

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.Client()
			if err != nil {
				return err
			}
			p := cli.Printer()

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

			headers := []string{"NAME", "EXECUTION MODE", "AGENTS", "ID"}
			rows := make([][]string, len(orgs))
			for i, o := range orgs {
				rows[i] = []string{o.Name, o.ExecutionMode, boolStr(o.AgentsEnabled), o.ID}
			}
			return p.Print(orgs, headers, rows)
		},
	}
}

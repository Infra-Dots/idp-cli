package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage personal API tokens",
	}
	cmd.AddCommand(newTokenListCmd())
	cmd.AddCommand(newTokenCreateCmd())
	cmd.AddCommand(newTokenRevokeCmd())
	return cmd
}

func newTokenListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List your API tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			tokens, err := client.ListTokens()
			if err != nil {
				return err
			}

			if p.Quiet {
				for _, t := range tokens {
					p.PrintID(t.ID)
				}
				return nil
			}

			headers := []string{"ID", "DESCRIPTION", "CREATED"}
			rows := make([][]string, len(tokens))
			for i, t := range tokens {
				rows[i] = []string{t.ID, t.Description, t.CreatedAt}
			}
			return p.Print(tokens, headers, rows)
		},
	}
}

func newTokenCreateCmd() *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API token",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			token, err := client.CreateToken(description)
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(token.ID)
				return nil
			}

			headers := []string{"ID", "DESCRIPTION", "CREATED"}
			rows := [][]string{{token.ID, token.Description, token.CreatedAt}}
			return p.Print(token, headers, rows)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Token description")
	_ = cmd.MarkFlagRequired("description")
	return cmd
}

func newTokenRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <token-id>",
		Short: "Revoke an API token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			if err := client.RevokeToken(args[0]); err != nil {
				return err
			}
			fmt.Printf("Token %s revoked\n", args[0])
			return nil
		},
	}
}

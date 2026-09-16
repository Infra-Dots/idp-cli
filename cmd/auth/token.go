package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
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
			client, err := cli.Client()
			if err != nil {
				return err
			}
			p := cli.Printer()

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

			headers := []string{"ID", "DESCRIPTION", "CREATED", "LAST USED", "EXPIRES"}
			rows := make([][]string, len(tokens))
			for i, t := range tokens {
				rows[i] = []string{t.ID, t.Description, t.Created, t.LastUsed, t.Expiration}
			}
			return p.Print(tokens, headers, rows)
		},
	}
}

func newTokenCreateCmd() *cobra.Command {
	var description string
	var expirationDays int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new API token",
		Long: `Create a new personal API token.

The token value is shown once, here, and cannot be retrieved again — store it
somewhere safe before closing your terminal.`,
		Example: `  idp auth token create --description "ci-deploy"
  idp auth token create --description "laptop" --expiration 90
  idp auth token create --description "ci" --quiet   # print only the token, for piping`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.Client()
			if err != nil {
				return err
			}
			p := cli.Printer()

			token, err := client.CreateToken(api.CreateTokenInput{
				Description: description,
				Expiration:  expirationDays,
			})
			if err != nil {
				return err
			}

			// --quiet prints the secret alone so it can be captured directly:
			//   export INFRADOTS_TOKEN=$(idp auth token create -d ci -q)
			if p.Quiet {
				p.PrintID(token.Token)
				return nil
			}

			headers := []string{"FIELD", "VALUE"}
			rows := [][]string{
				{"id", token.ID},
				{"description", token.Description},
				{"expiration", token.Expiration},
				{"token", token.Token},
			}
			if err := p.Print(token, headers, rows); err != nil {
				return err
			}
			// json/yaml already carry the token; only the table needs the warning,
			// and it goes to stderr so it never pollutes a piped value.
			fmt.Fprintln(os.Stderr, "\nSave this token now — it will not be shown again.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Token description")
	cmd.Flags().IntVar(&expirationDays, "expiration", 0, "Days until the token expires (default: 3650)")
	_ = cmd.MarkFlagRequired("description")
	return cmd
}

func newTokenRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <token-id>",
		Short: "Revoke an API token",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.Client()
			if err != nil {
				return err
			}
			if err := client.RevokeToken(args[0]); err != nil {
				return err
			}
			fmt.Printf("Token %s revoked\n", args[0])
			return nil
		},
	}
}

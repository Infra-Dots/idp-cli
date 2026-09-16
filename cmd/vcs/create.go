package vcs

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
)

func newCreateCmd() *cobra.Command {
	var name, description, vcsType, clientID, clientSecret, apiURL string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VCS connection",
		Long: `Create a new VCS connection.

InfraDots authenticates to providers over OAuth, so a connection is created from
an OAuth app's client credentials — not a personal access token. The connection
is created in "pending" status and only becomes usable once you complete the
OAuth callback in the browser; this command prints the URL to visit.`,
		Example: `  idp vcs create --org my-org --name github-main --type github \
    --client-id Iv1.abc123 --client-secret <secret>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			// The API requires a non-blank description; default it rather than
			// making the user invent one.
			if description == "" {
				description = name
			}

			v, err := scope.Client.CreateVCS(scope.Org, api.CreateVCSInput{
				Name:         name,
				Description:  description,
				VCSType:      vcsType,
				ClientID:     clientID,
				ClientSecret: clientSecret,
				APIURL:       apiURL,
			})
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				scope.Printer.PrintID(v.ID)
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "STATUS", "CREATED"}
			rows := [][]string{{v.ID, v.Name, v.VCSType, v.Status, v.CreatedDate}}
			if err := scope.Printer.Print(v, headers, rows); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr,
				"\nConnection created in %q status. Authorize it in the InfraDots web app\n"+
					"(Settings → VCS) to complete the OAuth handshake before using it in a workspace.\n",
				v.Status)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Connection name")
	cmd.Flags().StringVar(&description, "description", "", "Connection description")
	cmd.Flags().StringVar(&vcsType, "type", "", "VCS type: github, gitlab, bitbucket")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth app client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth app client secret")
	cmd.Flags().StringVar(&apiURL, "api-url", "", "Provider API URL (defaults per provider; set for self-hosted)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("client-secret")
	return cmd
}

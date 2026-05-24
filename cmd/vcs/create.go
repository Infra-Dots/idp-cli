package vcs

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newCreateCmd() *cobra.Command {
	var name, vcsType, token string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new VCS connection",
		Example: `  idp vcs create --org my-org --name github-main --type github --token ghp_xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			p := output.New(viper.GetString("output"), viper.GetBool("quiet"))

			v, err := client.CreateVCS(orgName, api.CreateVCSInput{
				Name:    name,
				VCSType: vcsType,
				Token:   token,
			})
			if err != nil {
				return err
			}

			if p.Quiet {
				p.PrintID(v.ID)
				return nil
			}

			headers := []string{"ID", "NAME", "TYPE", "CREATED"}
			rows := [][]string{{v.ID, v.Name, v.VCSType, v.CreatedAt}}
			return p.Print(v, headers, rows)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Connection name")
	cmd.Flags().StringVar(&vcsType, "type", "", "VCS type: github, gitlab, bitbucket")
	cmd.Flags().StringVar(&token, "token", "", "Personal access token")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	_ = cmd.MarkFlagRequired("token")
	return cmd
}

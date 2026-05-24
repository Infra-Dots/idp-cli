package job

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newOutputCmd() *cobra.Command {
	var wsName, stage string

	cmd := &cobra.Command{
		Use:   "output <job-id>",
		Short: "Print the output log of a job stage",
		Example: `  idp job output <job-id> --org my-org --workspace prod-infra
  idp job output <job-id> --org my-org --workspace prod-infra --stage apply`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			if wsName == "" {
				return output.NewError("--workspace is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))

			out, err := client.GetJobStageOutput(orgName, wsName, args[0], stage)
			if err != nil {
				return err
			}
			fmt.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	cmd.Flags().StringVar(&stage, "stage", "plan", "Stage to fetch: init, plan, apply")
	return cmd
}

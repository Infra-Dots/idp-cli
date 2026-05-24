package job

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

func newApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <job-id>",
		Short: "Approve a job to proceed with apply",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			if err := client.ApproveJob(orgName, args[0]); err != nil {
				return err
			}
			fmt.Printf("Job %s approved\n", args[0])
			return nil
		},
	}
}

func newCancelCmd() *cobra.Command {
	var wsName string

	cmd := &cobra.Command{
		Use:   "cancel <job-id>",
		Short: "Cancel a running job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			if wsName == "" {
				return output.NewError("--workspace is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			if err := client.CancelJob(orgName, wsName, args[0]); err != nil {
				return err
			}
			fmt.Printf("Job %s cancelled\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name")
	return cmd
}

func newDiscardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "discard <job-id>",
		Short: "Discard a job waiting for approval",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			orgName := viper.GetString("org")
			if orgName == "" {
				return output.NewError("--org is required")
			}
			client := api.NewClient(viper.GetString("host"), viper.GetString("token"))
			if err := client.DiscardJob(orgName, args[0]); err != nil {
				return err
			}
			fmt.Printf("Job %s discarded\n", args[0])
			return nil
		},
	}
}

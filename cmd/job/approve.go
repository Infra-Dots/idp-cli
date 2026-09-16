package job

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/cli"
)

func newApproveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "approve <job-id>",
		Short: "Approve a job to proceed with apply",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := scope.Client.ApproveJob(scope.Org, args[0]); err != nil {
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
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := cli.Require("workspace", wsName); err != nil {
				return err
			}
			if err := scope.Client.CancelJob(scope.Org, wsName, args[0]); err != nil {
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
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if err := scope.Client.DiscardJob(scope.Org, args[0]); err != nil {
				return err
			}
			fmt.Printf("Job %s discarded\n", args[0])
			return nil
		},
	}
}

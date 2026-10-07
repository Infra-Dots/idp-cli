package agent

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/localagent"
	"github.com/infradots/idp-cli/internal/output"
)

func newReviewCmd() *cobra.Command {
	var (
		opts    localagent.ReviewOptions
		offline bool
		native  bool
		image   string
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "review [PATH]",
		Short: "Review your working tree's change with the InfraDots agent (offline: free, your own model key)",
		Long: `Review the change on your working tree before you push: everything since the branch left --base,
committed or not. With --offline it runs on your machine (the idp-agent Docker image, or an installed
idp-agent with --native) with your own model key, and needs no InfraDots account.

The model account comes from your environment: ANTHROPIC_API_KEY, or MODEL_PROVIDER=bedrock|vertex|foundry
with MODEL_MAP and that provider's credentials.

Exit codes: 0 approved, 1 changes requested, 2 error.`,
		Example: `  idp agent review --offline
  idp agent review --offline envs/prod --plan plan.json
  idp agent review --offline -f json > review.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !offline {
				return output.NewError("only --offline reviews are available for now: they run locally with your own model key")
			}
			opts.Path = "."
			if len(args) == 1 {
				opts.Path = args[0]
			}
			if viper.GetString("output") == "json" {
				opts.Format = "json"
			}
			launcher := localagent.NewLauncher(native, image)
			program, programArgs, err := launcher.ReviewArgs(opts)
			if err != nil {
				return err
			}
			if dryRun {
				_, _ = fmt.Fprintln(os.Stdout, localagent.Redacted(program, programArgs))
				return nil
			}
			code, err := localagent.Run(program, programArgs)
			if err != nil {
				return err
			}
			if code != 0 {
				os.Exit(code) // the agent's verdict: 1 changes requested, 2 error
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&offline, "offline", false, "Review locally with your own model key, without InfraDots (required for now)")
	f.StringVar(&opts.Base, "base", "", "Branch the change is compared with (default: the remote's default branch)")
	f.StringVar(&opts.Plan, "plan", "", "A plan: `terraform show -json <planfile>` output, or the text of `terraform plan`")
	f.StringVar(&opts.Guidance, "guidance", "", "Team guidance file (default: .infradots/guidance.md, if present)")
	f.StringVar(&opts.Model, "model", "", "Model (default: the platform's review model)")
	f.StringVar(&opts.Instructions, "instructions", "", "Extra instructions for this review, e.g. \"focus on IAM\"")
	f.StringVar(&opts.IaC, "iac", "", "terraform or tofu, for the fmt check (default: terraform)")
	f.BoolVar(&native, "native", false, "Run an installed idp-agent instead of the Docker image")
	f.StringVar(&image, "image", "", "Agent image (default: "+localagent.DefaultImage+", or $"+localagent.ImageEnv+")")
	f.BoolVar(&dryRun, "dry-run", false, "Print the command that would run, and exit")
	return cmd
}

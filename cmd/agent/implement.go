package agent

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/localagent"
	"github.com/infradots/idp-cli/internal/output"
)

func newImplementCmd() *cobra.Command {
	var (
		opts    localagent.ImplementOptions
		offline bool
		native  bool
		image   string
		dryRun  bool
	)

	cmd := &cobra.Command{
		Use:   "implement REQUEST [PATH]",
		Short: "Implement a change on a new local branch with the InfraDots agent (offline: free, your own model key)",
		Long: `Ask for an infrastructure change and get it as a new branch of your repository. With --offline it runs on
your machine (the idp-agent Docker image, or an installed idp-agent with --native) with your own model key, and
needs no InfraDots account.

The agent works in a git worktree on a new branch, idp-agent/<timestamp>-<random>, cut from your last commit:
your checkout and anything uncommitted in it are never touched. It commits as you (your git user.name and
user.email) and never pushes. Review the branch, then merge it, push it or delete it.

The model account comes from your environment: ANTHROPIC_API_KEY, or MODEL_PROVIDER=bedrock|vertex|foundry
with MODEL_MAP and that provider's credentials.

Exit codes: 0 a branch was made, 1 nothing changed, 2 error.`,
		Example: `  idp agent implement --offline "add versioning to the logs bucket"
  idp agent implement --offline "add a read replica" envs/prod --patch
  idp agent implement --offline "tag everything with team=platform" -f json`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !offline {
				return output.NewError("only --offline implementations are available here for now: they run locally " +
					"with your own model key (implementations on your workspaces run on InfraDots)")
			}
			opts.Request = args[0]
			opts.Path = "."
			if len(args) == 2 {
				opts.Path = args[1]
			}
			if viper.GetString("output") == "json" {
				opts.Format = "json"
			}
			launcher := localagent.NewLauncher(native, image)
			program, programArgs, err := launcher.ImplementArgs(opts)
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
				os.Exit(code) // the agent's outcome: 1 nothing changed, 2 error
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&offline, "offline", false, "Implement locally with your own model key, without InfraDots (required for now)")
	f.BoolVar(&opts.Patch, "patch", false, "Include the full diff in the report")
	f.StringVar(&opts.Model, "model", "", "Model (default: the platform's implement model)")
	f.StringVar(&opts.IaC, "iac", "", "terraform or tofu (default: terraform)")
	f.BoolVar(&native, "native", false, "Run an installed idp-agent instead of the Docker image")
	f.StringVar(&image, "image", "", "Agent image (default: "+localagent.DefaultImage+", or $"+localagent.ImageEnv+")")
	f.BoolVar(&dryRun, "dry-run", false, "Print the command that would run, and exit")
	return cmd
}

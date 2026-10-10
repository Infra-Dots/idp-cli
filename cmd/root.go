package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/cmd/agent"
	"github.com/infradots/idp-cli/cmd/auth"
	"github.com/infradots/idp-cli/cmd/job"
	"github.com/infradots/idp-cli/cmd/org"
	"github.com/infradots/idp-cli/cmd/sshkey"
	"github.com/infradots/idp-cli/cmd/variable"
	"github.com/infradots/idp-cli/cmd/vcs"
	"github.com/infradots/idp-cli/cmd/workspace"
	"github.com/infradots/idp-cli/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "idp",
	Short: "InfraDots CLI — manage your IaC platform from the terminal",
	Long: `idp is the official command-line interface for InfraDots.

It lets you manage organizations, workspaces, jobs, variables, VCS connections,
and AI agent history directly from your terminal or CI pipelines.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("host", "", "InfraDots API host (overrides config)")
	rootCmd.PersistentFlags().String("token", "", "API token (overrides config)")
	rootCmd.PersistentFlags().StringP("org", "o", "", "Organization name")
	rootCmd.PersistentFlags().String("profile", "", "Config profile to use (default: active profile)")
	rootCmd.PersistentFlags().StringP("output", "f", "table", "Output format: table, json, yaml")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Print only IDs/names (pipe-friendly)")

	_ = viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	_ = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	_ = viper.BindPFlag("org", rootCmd.PersistentFlags().Lookup("org"))
	_ = viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))

	viper.AutomaticEnv()
	viper.SetEnvPrefix("INFRADOTS")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(auth.NewCmd())
	rootCmd.AddCommand(org.NewCmd())
	rootCmd.AddCommand(workspace.NewCmd())
	rootCmd.AddCommand(job.NewCmd())
	rootCmd.AddCommand(variable.NewCmd())
	rootCmd.AddCommand(vcs.NewCmd())
	rootCmd.AddCommand(sshkey.NewCmd())
	rootCmd.AddCommand(agent.NewCmd())
}

// initConfig loads the config file and merges profile values into viper,
// unless --host / --token were explicitly passed on the command line.
func initConfig() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
		return
	}

	profileName := viper.GetString("profile")
	p := cfg.ActiveProfile(profileName)

	// Only apply profile values if the flag was not explicitly set.
	if viper.GetString("host") == "" && p.Host != "" {
		viper.Set("host", p.Host)
	}
	if viper.GetString("token") == "" && p.Token != "" {
		viper.Set("token", p.Token)
	}
	if viper.GetString("org") == "" && p.DefaultOrg != "" {
		viper.Set("org", p.DefaultOrg)
	}
}

// Client, Printer and RequireOrg live in internal/cli so the command
// subpackages can reach them — this package imports them, so they cannot
// import it back.

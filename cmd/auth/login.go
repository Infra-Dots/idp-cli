package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/infradots/idp-cli/internal/config"
	"github.com/infradots/idp-cli/internal/output"
)

func newLoginCmd() *cobra.Command {
	var host, token, profileName string
	var noPrompt bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save API credentials to a config profile",
		Example: `  idp auth login
  idp auth login --host https://api.infradots.com --token mytoken --no-prompt
  idp auth login --profile local --host http://localhost:8000`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			if profileName == "" {
				profileName = cfg.DefaultProfile
			}
			if profileName == "" {
				profileName = "default"
			}

			// Interactive prompts when flags not supplied.
			if !noPrompt {
				if host == "" {
					host = prompt("Host", "https://api.infradots.com")
				}
				if token == "" {
					token = promptSecret("Token")
				}
			}

			if host == "" {
				return fmt.Errorf("--host is required")
			}
			if token == "" {
				return fmt.Errorf("--token is required")
			}

			existing := cfg.ActiveProfile(profileName)
			existing.Host = host
			existing.Token = token
			cfg.SetProfile(profileName, existing)

			if cfg.DefaultProfile == "" {
				cfg.DefaultProfile = profileName
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			fmt.Fprintf(os.Stdout, "Logged in. Profile %q saved to %s\n", profileName, config.ConfigPath())
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "InfraDots API host")
	cmd.Flags().StringVar(&token, "token", "", "API token")
	cmd.Flags().StringVar(&profileName, "profile", "", "Profile name to save credentials under (default: \"default\")")
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "Disable interactive prompts (requires --host and --token)")

	return cmd
}

func newLogoutCmd() *cobra.Command {
	var profileName string

	return &cobra.Command{
		Use:   "logout",
		Short: "Remove the token from a config profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if profileName == "" {
				profileName = cfg.DefaultProfile
			}
			if profileName == "" {
				profileName = "default"
			}
			cfg.RemoveToken(profileName)
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "Token removed from profile %q\n", profileName)
			return nil
		},
	}
}

func prompt(label, placeholder string) string {
	fmt.Printf("%s [%s]: ", label, placeholder)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return placeholder
	}
	return line
}

func promptSecret(label string) string {
	fmt.Printf("%s: ", label)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		output.Fatal("reading token: %v", err)
	}
	return strings.TrimSpace(string(b))
}

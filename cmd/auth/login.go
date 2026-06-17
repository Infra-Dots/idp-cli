package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/browser"
	"github.com/infradots/idp-cli/internal/config"
)

func newLoginCmd() *cobra.Command {
	var host, token, appURL, profileName string
	var noPrompt bool

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Sign in via the browser and save API credentials to a config profile",
		Long: `Sign in to InfraDots.

By default this opens your browser, authenticates you in the InfraDots web app,
and saves a freshly minted API token to your config profile.

For CI or headless use, pass --token to skip the browser and store a token you
created manually in the web app under Settings → Tokens.`,
		Example: `  idp auth login
  idp auth login --host http://localhost:8000 --app-url http://localhost:3001
  idp auth login --token <token> --no-prompt   # CI / headless`,
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

			// A token passed explicitly means manual/CI mode; otherwise we run
			// the interactive browser login.
			useBrowser := token == ""

			if !noPrompt {
				if host == "" {
					host = prompt("Host", "https://api.infradots.com")
				}
				if useBrowser && appURL == "" {
					appURL = prompt("Web app URL", defaultAppURL(host))
				}
			}

			if host == "" {
				return fmt.Errorf("--host is required")
			}

			if useBrowser {
				if appURL == "" {
					appURL = defaultAppURL(host)
				}
				token, err = browser.Login(appURL)
				if err != nil {
					return fmt.Errorf("browser login failed: %w", err)
				}
			}

			if token == "" {
				return fmt.Errorf("--token is required (or omit --no-prompt to sign in via the browser)")
			}

			existing := cfg.ActiveProfile(profileName)
			existing.Host = host
			existing.Token = token
			if appURL != "" {
				existing.WebURL = appURL
			}
			cfg.SetProfile(profileName, existing)

			if cfg.DefaultProfile == "" {
				cfg.DefaultProfile = profileName
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			_, _ = fmt.Fprintf(os.Stdout, "Logged in. Profile %q saved to %s\n", profileName, config.ConfigPath())
			return nil
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "InfraDots API host")
	cmd.Flags().StringVar(&token, "token", "", "API token (skips browser login; for CI/headless use)")
	cmd.Flags().StringVar(&appURL, "app-url", "", "InfraDots web app URL for browser login (default derived from --host)")
	cmd.Flags().StringVar(&profileName, "profile", "", "Profile name to save credentials under (default: \"default\")")
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "Disable interactive prompts (requires --token, or --host with browser login)")

	return cmd
}

// defaultAppURL guesses the web app origin from the API host so the common
// cases (local dev and production) work without an explicit --app-url.
func defaultAppURL(host string) string {
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		return "http://localhost:3001"
	}
	return "https://app.infradots.com"
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
			_, _ = fmt.Fprintf(os.Stdout, "Token removed from profile %q\n", profileName)
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

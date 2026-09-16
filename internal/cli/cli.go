// Package cli resolves the flags and config that nearly every command needs
// into an API client, a printer, and an organization name.
//
// These helpers used to live in package cmd, but cmd imports every command
// subpackage, so no subpackage could import them back — they were unreachable,
// and each command re-implemented them inline. Keeping them here breaks that
// cycle: both cmd and cmd/* can import internal/cli.
package cli

import (
	"github.com/spf13/viper"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/output"
)

// Client builds an API client from the resolved host/token, with an error that
// tells the user how to fix a missing credential rather than failing later with
// an opaque URL error.
func Client() (*api.Client, error) {
	host := viper.GetString("host")
	token := viper.GetString("token")

	if host == "" {
		return nil, output.NewError("no host configured — run `idp auth login` or pass --host")
	}
	if token == "" {
		return nil, output.NewError("no token configured — run `idp auth login` or pass --token")
	}
	return api.NewClient(host, token), nil
}

// Printer builds an output printer from the resolved --output/--quiet flags.
func Printer() *output.Printer {
	return output.New(viper.GetString("output"), viper.GetBool("quiet"))
}

// RequireOrg returns the resolved organization name, or an error naming both
// ways to supply one.
func RequireOrg() (string, error) {
	org := viper.GetString("org")
	if org == "" {
		return "", output.NewError("--org is required (or set default_org in your config profile)")
	}
	return org, nil
}

// Require returns an error when a mandatory flag was left empty.
func Require(flag, value string) error {
	if value == "" {
		return output.NewError("--%s is required", flag)
	}
	return nil
}

// OrgScope bundles the client, printer, and organization that org-scoped
// commands all need.
type OrgScope struct {
	Client  *api.Client
	Printer *output.Printer
	Org     string
}

// NewOrgScope resolves all three, reporting the first problem it finds.
func NewOrgScope() (*OrgScope, error) {
	org, err := RequireOrg()
	if err != nil {
		return nil, err
	}
	client, err := Client()
	if err != nil {
		return nil, err
	}
	return &OrgScope{Client: client, Printer: Printer(), Org: org}, nil
}

// Package sshkey is `idp ssh-key`: an organization's SSH keys for module sources fetched over SSH.
package sshkey

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
)

// NewCmd returns the `idp ssh-key` command group.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh-key",
		Short: "Manage SSH keys for private module sources",
		Long: `Manage the organization's SSH keys for module sources fetched over SSH
(git::ssh://git@github.com/acme/modules.git//vpc, git@github.com:acme/modules.git).

A workspace uses one with ` + "`idp workspace update <ws> --ssh-key <name>`" + `; each of its jobs hands it to the
executor for the run. InfraDots never returns a private key: only its public half and fingerprint.`,
	}
	cmd.AddCommand(newListCmd(), newCreateCmd(), newUpdateCmd(), newDeleteCmd())
	return cmd
}

var listHeaders = []string{"ID", "NAME", "FINGERPRINT", "WORKSPACES"}

func listRow(k api.SSHKey) []string {
	return []string{k.ID, k.Name, k.Fingerprint, strings.Join(k.Workspaces, ",")}
}

var detailHeaders = []string{"FIELD", "VALUE"}

func detailRows(k *api.SSHKey) [][]string {
	return [][]string{
		{"name", k.Name},
		{"id", k.ID},
		{"fingerprint", k.Fingerprint},
		{"public_key", k.PublicKey},
		{"workspaces", strings.Join(k.Workspaces, ",")},
	}
}

// readFile reads a key or known_hosts file; "-" is stdin. Keys come from files, never flag values, so they
// stay out of shell history and the process list.
func readFile(path string) (string, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(data), nil
}

package sshkey

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
)

func newCreateCmd() *cobra.Command {
	var name, privateKeyFile, knownHostsFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add an SSH key for private module sources",
		Long: `Add an SSH key for module sources fetched over SSH.

The key must have no passphrase (an executor can't enter one); generate one for InfraDots with
  ssh-keygen -t ed25519 -N "" -C infradots-modules -f infradots_modules
then add the printed public key as a read-only deploy key on the module repositories.

Executors already trust github.com, gitlab.com and bitbucket.org; for other SSH hosts pass their
known_hosts lines with --known-hosts-file (e.g. from ` + "`ssh-keyscan git.acme.internal`" + `).`,
		Example: `  idp ssh-key create --org my-org --name private-modules --private-key-file infradots_modules
  idp ssh-key create --name internal-git --private-key-file key --known-hosts-file known_hosts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			in := api.SSHKeyInput{Name: name}
			if in.PrivateKey, err = readFile(privateKeyFile); err != nil {
				return err
			}
			if knownHostsFile != "" {
				knownHosts, err := readFile(knownHostsFile)
				if err != nil {
					return err
				}
				in.KnownHosts = &knownHosts
			}
			key, err := scope.Client.CreateSSHKey(scope.Org, in)
			if err != nil {
				return err
			}
			if scope.Printer.Quiet {
				scope.Printer.PrintID(key.ID)
				return nil
			}
			if err := scope.Printer.Print(key, detailHeaders, detailRows(key)); err != nil {
				return err
			}
			fmt.Fprintf(os.Stderr, "\nAdd the public key as a deploy key on the module repositories, then use it in a "+
				"workspace:\n  idp workspace update <workspace> --ssh-key %s\n", key.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Key name, unique in the organization")
	cmd.Flags().StringVar(&privateKeyFile, "private-key-file", "", "Private key file (OpenSSH or PEM, no passphrase); - for stdin")
	cmd.Flags().StringVar(&knownHostsFile, "known-hosts-file", "", "known_hosts lines for hosts other than GitHub, GitLab and Bitbucket")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("private-key-file")
	return cmd
}

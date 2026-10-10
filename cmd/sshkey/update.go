package sshkey

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

func newUpdateCmd() *cobra.Command {
	var name, privateKeyFile, knownHostsFile string

	cmd := &cobra.Command{
		Use:   "update <name|id>",
		Short: "Rename an SSH key, rotate it, or replace its known_hosts",
		Example: `  idp ssh-key update private-modules --private-key-file new_key
  idp ssh-key update internal-git --known-hosts-file known_hosts`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}
			if name == "" && privateKeyFile == "" && !cmd.Flags().Changed("known-hosts-file") {
				return output.NewError("nothing to update: pass --name, --private-key-file or --known-hosts-file")
			}
			key, err := scope.Client.ResolveSSHKey(scope.Org, args[0])
			if err != nil {
				return err
			}
			in := api.SSHKeyInput{Name: name}
			if privateKeyFile != "" {
				if in.PrivateKey, err = readFile(privateKeyFile); err != nil {
					return err
				}
			}
			if cmd.Flags().Changed("known-hosts-file") {
				knownHosts := ""
				if knownHostsFile != "" {
					if knownHosts, err = readFile(knownHostsFile); err != nil {
						return err
					}
				}
				in.KnownHosts = &knownHosts
			}
			updated, err := scope.Client.UpdateSSHKey(scope.Org, key.ID, in)
			if err != nil {
				return err
			}
			if scope.Printer.Quiet {
				scope.Printer.PrintID(updated.ID)
				return nil
			}
			return scope.Printer.Print(updated, detailHeaders, detailRows(updated))
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New name")
	cmd.Flags().StringVar(&privateKeyFile, "private-key-file", "", "New private key file (rotates the key); - for stdin")
	cmd.Flags().StringVar(&knownHostsFile, "known-hosts-file", "", `Replace the known_hosts lines ("" clears them)`)
	return cmd
}

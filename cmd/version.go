package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the idp-cli version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("idp-cli %s\n", Version)
	},
}

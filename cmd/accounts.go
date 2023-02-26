package cmd

import (
	"github.com/spf13/cobra"
)

var apiAddr string

func init() {
	rootCmd.AddCommand(accountsCmd)
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage gpodder2go user accounts",
}

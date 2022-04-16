package cmd

import (
	"github.com/spf13/cobra"
)

var apiAddr string

func init() {
	rootCmd.AddCommand(accountsCmd)
	rootCmd.PersistentFlags().StringVarP(&apiAddr, "api-addr", "", "localhost:3005", "api addr")
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "Manage gpodder2go user accounts",
}

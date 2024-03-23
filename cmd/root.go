package cmd

import (
	"github.com/spf13/cobra"
)

var database string

var rootCmd = &cobra.Command{
	Use:   "gpodder2go",
	Short: "gpodder2go is a drop-in replacement for gpodder/mygpo server",
	Long:  `A simple self-hosted, golang, drop-in replacement for gpodder/mygpo server to handle podcast subscriptions management for gpodder clients`,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&database, "database", "d", "g2g.db", "filename of sqlite3 database to use")
}

func Execute() error {
	return rootCmd.Execute()
}

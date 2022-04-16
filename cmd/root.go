package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gpodder2go",
	Short: "gpodder2go is a drop-in replacement for gpodder/mygpo server",
	Long:  `A simple self-hosted, golang, drop-in replacement for gpodder/mygpo server to handle podcast subscriptions management for gpodder clients`,
}

func Execute() error {
	return rootCmd.Execute()
}

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(devicesCmd)
}

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices",
}

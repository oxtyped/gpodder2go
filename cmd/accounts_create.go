package cmd

import (
	"github.com/spf13/cobra"
)

var password string

func init() {
	accountsCmd.AddCommand(accountsCreateCmd)
	accountsCreateCmd.Flags().StringVarP(&password, "password", "p", "", "Password to use for user (required)")
	accountsCreateCmd.MarkFlagRequired("password")

}

var accountsCreateCmd = &cobra.Command{
	Use:   "create [username]",
	Short: "Create user account",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

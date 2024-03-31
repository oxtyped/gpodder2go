package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/oxtyped/gpodder2go/pkg/data"
)

var password, email, name string

func init() {
	accountsCmd.AddCommand(accountsCreateCmd)
	accountsCreateCmd.Flags().StringVarP(&password, "password", "p", "", "Password to use for user (required)")
	accountsCreateCmd.Flags().StringVarP(&email, "email", "e", "", "User email")
	accountsCreateCmd.Flags().StringVarP(&name, "name", "n", "", "User Display Name")
	accountsCreateCmd.MarkFlagRequired("password")
	accountsCreateCmd.MarkFlagRequired("email")
	accountsCreateCmd.MarkFlagRequired("name")
}

var accountsCreateCmd = &cobra.Command{
	Use:   "create [username]",
	Short: "Create user account",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]

		dataInterface := data.NewSQLite(database)

		err := dataInterface.AddUser(username, password, email, name)
		if err != nil {
			log.Fatalf("could not create user: %#v", err)
		}

		log.Printf("😍 User %s created!", username)
	},
}

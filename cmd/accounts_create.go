package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
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
		// make apiAddr separate it out into username/port
		val := url.Values{"username": {username}, "password": {"password"}, "email": {email}, "name": {name}}

		addr := fmt.Sprintf("http://%s/api/internal/users", apiAddr)
		res, err := http.PostForm(addr, val)
		if err != nil {
			log.Printf("error reaching API server: %#v", err.Error())
			return
		}

		if res.StatusCode != 201 {
			log.Println("Could not create status code")
		}

		log.Printf("😍 User %s created!", username)
		return

	},
}

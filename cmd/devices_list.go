package cmd

import (
	"fmt"
	"log"

	"github.com/oxtyped/gpodder2go/pkg/data"
	"github.com/spf13/cobra"
)

func init() {
	devicesCmd.AddCommand(devicesListCmd)
}

var devicesListCmd = &cobra.Command{
	Use:   "list [username]",
	Short: "Get all devices belong to username",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username := args[0]

		if username == "" {
			log.Fatalln("expected arguments to have 1")
		}

		dataInterface := data.NewSQLite(database)
		devices, err := dataInterface.GetDevicesFromUsername(username)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("devices are : %#v", devices)
		// make apiAddr separate it out into username/port
		return

	},
}

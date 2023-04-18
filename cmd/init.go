package cmd

import (
	"database/sql"
	"embed"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/cobra"
)

//go:embed migrations/*.sql
var fs embed.FS

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Setup the necessary environments",
	Run: func(cmd *cobra.Command, args []string) {

		// create sqlite file
		// run migration file
		db, err := sql.Open("sqlite", database)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		instance, err := sqlite.WithInstance(db, &sqlite.Config{})
		if err != nil {
			log.Fatal(err)
		}

		src, err := iofs.New(fs, "migrations")
		if err != nil {
			log.Fatal(err)
		}

		m, err := migrate.NewWithInstance("iofs", src, "sqlite", instance)
		if err != nil {
			log.Fatal(err)
		}

		// modify for Down
		if err := m.Up(); err != nil {
			log.Fatal(err)
		}
	},
}

package cmd

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

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

		fSrc, err := (&file.File{}).Open("./migrations")
		if err != nil {
			log.Fatal(err)
		}

		m, err := migrate.NewWithInstance("file", fSrc, "sqlite", instance)
		if err != nil {
			log.Fatal(err)
		}

		// modify for Down
		if err := m.Up(); err != nil {
			log.Fatal(err)
		}
	},
}

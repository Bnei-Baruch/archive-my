package cmd

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Migration DB from files",
	Run: func(cmd *cobra.Command, args []string) {
		MigrationUpFn(cmd, args)
		MigrationDownFn(cmd, args)
		MigrationUpFn(cmd, args)
	},
}

var migrationUpCmd = &cobra.Command{
	Use:   "up",
	Short: "up migration data base",
	Run:   MigrationUpFn,
}

var migrationDownCmd = &cobra.Command{
	Use:   "down",
	Short: "down migration data base",
	Run:   MigrationDownFn,
}

func init() {
	rootCmd.AddCommand(migrationCmd)
	migrationCmd.AddCommand(migrationUpCmd)
	migrationCmd.AddCommand(migrationDownCmd)
}

func MigrationUpFn(cmd *cobra.Command, args []string) {
	m, err := migrate.New(viper.GetString("app.migration-dir"), viper.GetString("app.mydb"))
	if err != nil {
		log.Panic(err)
	}
	if err := m.Up(); err != nil && err.Error() != "no change" {
		log.Panic(err)
	}
}

func MigrationDownFn(cmd *cobra.Command, args []string) {
	m, err := migrate.New(viper.GetString("app.migration-dir"), viper.GetString("app.mydb"))
	if err != nil {
		log.Panic(err)
	}
	if err := m.Up(); err != nil && err.Error() != "no change" {
		log.Panic(err)
	}
}

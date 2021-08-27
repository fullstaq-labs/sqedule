package main

import (
	"context"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/server/dbmigrations"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// dbRollbackCcmd represents the 'db rollback' command
var dbRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := viper.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}

		err = dbRollbackCmd_checkConfig(viper.GetViper())
		if err != nil {
			return err
		}

		dbLogger, err := createLoggerWithLevel(viper.GetString("db-log-level"))
		if err != nil {
			return fmt.Errorf("Error initializing logger: %w", err)
		}

		db, err := dbutils.EstablishDatabaseConnection(
			viper.GetString("db-type"),
			viper.GetString("db-connection"),
			&gorm.Config{
				Logger: dbLogger,
			})
		if err != nil {
			return fmt.Errorf("Error establishing database connection: %w", err)
		}

		gormigrateOptions := createGormigrateOptions(logger)
		migrator := gormigrate.New(db, &gormigrateOptions, dbmigrations.DbMigrations())
		if err := migrator.RollbackTo(viper.GetString("target")); err != nil {
			return fmt.Errorf("Error rolling back database schema: %w", err)
		}

		logger.Info(context.Background(), "Database schema rollback complete")
		return nil
	},
}

func dbRollbackCmd_checkConfig(viper *viper.Viper) error {
	spec := cli.ConfigRequirementSpec{
		StringNonEmpty: []string{"target"},
	}
	defineDatabaseConnectionConfigRequirementSpec(&spec)
	return cli.RequireConfigOptions(viper, spec)
}

func init() {
	cmd := dbRollbackCmd
	flags := cmd.Flags()
	dbCmd.AddCommand(cmd)

	defineDatabaseConnectionFlags(cmd)

	flags.String("target", "", "migration ID to rollback to (required)")
}

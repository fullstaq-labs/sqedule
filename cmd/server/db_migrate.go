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
	gormlogger "gorm.io/gorm/logger"
)

// dbMigrateCcmd represents the 'db migrate' command
var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		err := dbMigrateCmd_checkConfig(viper.GetViper())
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

		if viper.GetBool("reset") {
			logger.Info(context.Background(), "Resetting database")
			if err := dbutils.ResetDatabase(context.Background(), db); err != nil {
				return fmt.Errorf("Error resetting database: %w", err)
			}
		}

		gormigrateOptions := createGormigrateOptions(logger)
		migrator := gormigrate.New(db, &gormigrateOptions, dbmigrations.DbMigrations())
		if len(viper.GetString("up-to")) > 0 {
			err = migrator.MigrateTo(viper.GetString("up-to"))
		} else {
			err = migrator.Migrate()
		}
		if err != nil {
			return fmt.Errorf("Error running migrations: %w", err)
		}

		logger.Info(context.Background(), "Database migration complete")
		return nil
	},
}

func dbMigrateCmd_checkConfig(viper *viper.Viper) error {
	spec := cli.ConfigRequirementSpec{}
	defineDatabaseConnectionConfigRequirementSpec(&spec)
	return cli.RequireConfigOptions(viper, spec)
}

func createGormigrateOptions(logger gormlogger.Interface) gormigrate.Options {
	return gormigrate.Options{
		TableName:                 "migrations",
		IDColumnName:              "id",
		IDColumnSize:              255,
		UseTransaction:            true,
		ValidateUnknownMigrations: true,
		Logger:                    logger,
	}
}

func init() {
	cmd := dbMigrateCmd
	flags := cmd.Flags()
	dbCmd.AddCommand(cmd)

	defineDatabaseConnectionFlags(cmd)

	flags.Bool("reset", false, "wipe the database and recreate schema from scratch (DANGER)")
	flags.String("up-to", "", "run migrations up to the given migration ID")
}

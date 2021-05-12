package main

import (
	"context"
	"fmt"

	"github.com/fullstaq-labs/sqedule/server/dbmigrations"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var dbMigrateFlags struct {
	dbconn databaseConnectionFlags
	reset  *bool
	upTo   *string
}

// dbMigrateCcmd represents the 'db migrate' command
var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbLogger, err := createLoggerWithLevel(*dbMigrateFlags.dbconn.dbLogLevel)
		if err != nil {
			return fmt.Errorf("Error initializing logger: %w", err)
		}

		db, err := dbutils.EstablishDatabaseConnection(
			*dbMigrateFlags.dbconn.dbType,
			*dbMigrateFlags.dbconn.dbConnection,
			&gorm.Config{
				Logger: dbLogger,
			})
		if err != nil {
			return fmt.Errorf("Error establishing database connection: %w", err)
		}

		if *dbMigrateFlags.reset {
			logger.Info(context.Background(), "Resetting database")
			if err := dbutils.ResetDatabase(context.Background(), db); err != nil {
				return fmt.Errorf("Error resetting database: %w", err)
			}
		}

		gormigrateOptions := createGormigrateOptions(logger)
		migrator := gormigrate.New(db, &gormigrateOptions, dbmigrations.DbMigrations())
		if len(*dbMigrateFlags.upTo) > 0 {
			err = migrator.MigrateTo(*dbMigrateFlags.upTo)
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

func init() {
	cmd := dbMigrateCmd
	flags := cmd.Flags()
	dbCmd.AddCommand(cmd)

	dbMigrateFlags.dbconn = defineDatabaseConnectionFlags(cmd)

	dbMigrateFlags.reset = flags.Bool("reset", false, "wipe the database and recreate schema from scratch (DANGER)")
	dbMigrateFlags.upTo = flags.String("up-to", "", "run migrations up to the given migration ID")
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

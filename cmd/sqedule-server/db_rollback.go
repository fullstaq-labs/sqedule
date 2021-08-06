package main

import (
	"context"
	"fmt"

	"github.com/fullstaq-labs/sqedule/server/dbmigrations"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var dbRollbackFlags struct {
	dbconn databaseConnectionFlags
	target *string
}

// dbRollbackCcmd represents the 'db rollback' command
var dbRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback database schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())

		dbLogger, err := createLoggerWithLevel(*dbRollbackFlags.dbconn.dbLogLevel)
		if err != nil {
			return fmt.Errorf("Error initializing logger: %w", err)
		}

		db, err := dbutils.EstablishDatabaseConnection(
			*dbRollbackFlags.dbconn.dbType,
			*dbRollbackFlags.dbconn.dbConnection,
			&gorm.Config{
				Logger: dbLogger,
			})
		if err != nil {
			return fmt.Errorf("Error establishing database connection: %w", err)
		}

		gormigrateOptions := createGormigrateOptions(logger)
		migrator := gormigrate.New(db, &gormigrateOptions, dbmigrations.DbMigrations())
		if err := migrator.RollbackTo(*dbRollbackFlags.target); err != nil {
			return fmt.Errorf("Error rolling back database schema: %w", err)
		}

		logger.Info(context.Background(), "Database schema rollback complete")
		return nil
	},
}

func init() {
	cmd := dbRollbackCmd
	flags := cmd.Flags()
	dbCmd.AddCommand(cmd)

	dbRollbackFlags.dbconn = defineDatabaseConnectionFlags(cmd)

	dbRollbackFlags.target = flags.String("target", "", "migration ID to rollback to (required)")
	cmd.MarkFlagRequired("target")
}

package main

import (
	"github.com/spf13/cobra"
)

// dbCmd represents the 'db' command
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage the database",
}

type databaseConnectionFlags struct {
	dbType       *string
	dbConnection *string
	dbLogLevel   *string
}

func defineDatabaseConnectionFlags(cmd *cobra.Command) (result databaseConnectionFlags) {
	result.dbType = cmd.Flags().StringP("db-type", "T", "", "database type (required). Learn more at 'sqedule db help'")
	result.dbConnection = cmd.Flags().StringP("db-connection", "D", "", "database connection string. Learn more at 'schedule db help'")
	result.dbLogLevel = cmd.Flags().String("db-log-level", "silent", "log level for database operations. One of: error,warn,info,silent")
	cmd.MarkFlagRequired("db-type")
	return result
}

func init() {
	rootCmd.AddCommand(dbCmd)
}

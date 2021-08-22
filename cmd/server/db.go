package main

import (
	"github.com/fullstaq-labs/sqedule/cli"
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
	result.dbType = cmd.Flags().StringP("db-type", "T", "", "database type (required). Learn more at https://fullstaq-labs.github.io/sqedule/server_guide/config/reference/")
	result.dbConnection = cmd.Flags().StringP("db-connection", "D", "", "database connection string. Learn more at https://fullstaq-labs.github.io/sqedule/server_guide/config/postgresql/")
	result.dbLogLevel = cmd.Flags().String("db-log-level", "silent", "log level for database operations. One of: error,warn,info,silent")
	return result
}

func defineDatabaseConnectionConfigRequirementSpec(spec *cli.ConfigRequirementSpec) {
	spec.StringNonEmpty = append(spec.StringNonEmpty, "db-type")
}

func init() {
	cmd := dbCmd
	rootCmd.AddCommand(cmd)
}

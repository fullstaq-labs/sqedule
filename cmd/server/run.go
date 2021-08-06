package main

import (
	"context"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/server/approvalrulesprocessing"
	"github.com/fullstaq-labs/sqedule/server/dbmigrations"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"github.com/fullstaq-labs/sqedule/server/httpapi"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	runDefaultBind = "localhost"
	runDefaultPort = 3001
)

// runCmd represents the 'run' command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Sqedule API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())
		err := runCmd_checkConfig(viper.GetViper())
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

		if viper.GetBool("auto-db-migrate") {
			logger.Warn(context.Background(), "Automatically migrating database schemas")
			gormigrateOptions := createGormigrateOptions(dbLogger)
			migrator := gormigrate.New(db, &gormigrateOptions, dbmigrations.DbMigrations())
			if err = migrator.Migrate(); err != nil {
				return fmt.Errorf("Error running migrations: %w", err)
			}
		}

		engine := gin.Default()
		ctx := httpapi.Context{
			Db:         db,
			CorsOrigin: viper.GetString("cors-origin"),
		}

		err = ctx.SetupRouter(engine)
		if err != nil {
			return fmt.Errorf("Error setting up router: %w", err)
		}

		err = approvalrulesprocessing.ProcessAllPendingReleasesInBackground(ctx.Db)
		if err != nil {
			return fmt.Errorf("Error processing pending releases in the background: %w", err)
		}

		engine.Run(fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")))
		return nil
	},
}

func runCmd_checkConfig(viper *viper.Viper) error {
	spec := cli.ConfigRequirementSpec{}
	defineDatabaseConnectionConfigRequirementSpec(&spec)
	return cli.RequireConfigOptions(viper, spec)
}

func init() {
	cmd := runCmd
	flags := cmd.Flags()
	rootCmd.AddCommand(cmd)

	defineDatabaseConnectionFlags(cmd)

	flags.String("bind", runDefaultBind, "IP to listen on")
	flags.Int("port", runDefaultPort, "port to listen on")
	flags.String("cors-origin", "", "CORS origin to allow")
	flags.Bool("auto-db-migrate", true, "automatically migrate database schema")
}

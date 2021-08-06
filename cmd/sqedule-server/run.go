package main

import (
	"context"
	"fmt"

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

var runFlags struct {
	dbconn     databaseConnectionFlags
	bind       *string
	port       *int
	corsOrigin *string
}

// runCmd represents the 'run' command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the Sqedule API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())

		dbLogger, err := createLoggerWithLevel(*runFlags.dbconn.dbLogLevel)
		if err != nil {
			return fmt.Errorf("Error initializing logger: %w", err)
		}

		db, err := dbutils.EstablishDatabaseConnection(
			*runFlags.dbconn.dbType,
			*runFlags.dbconn.dbConnection,
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
			CorsOrigin: *runFlags.corsOrigin,
		}

		err = ctx.SetupRouter(engine)
		if err != nil {
			return fmt.Errorf("Error setting up router: %w", err)
		}

		err = approvalrulesprocessing.ProcessAllPendingReleasesInBackground(ctx.Db)
		if err != nil {
			return fmt.Errorf("Error processing pending releases in the background: %w", err)
		}

		engine.Run(fmt.Sprintf("%s:%d", *runFlags.bind, *runFlags.port))
		return nil
	},
}

func init() {
	cmd := runCmd
	flags := cmd.Flags()
	rootCmd.AddCommand(cmd)

	runFlags.dbconn = defineDatabaseConnectionFlags(cmd)

	runFlags.bind = flags.String("bind", runDefaultBind, "IP to listen on")
	runFlags.port = flags.Int("port", runDefaultPort, "port to listen on")
	runFlags.corsOrigin = flags.String("cors-origin", "", "CORS origin to allow")
	flags.Bool("auto-db-migrate", true, "automatically migrate database schema")
}

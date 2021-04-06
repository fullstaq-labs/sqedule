package cmd

import (
	"fmt"

	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/fullstaq-labs/sqedule/httpapi"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

const (
	serverDefaultBind = "localhost"
	serverDefaultPort = 3001
)

var serverFlags struct {
	dbconn     databaseConnectionFlags
	bind       *string
	port       *int
	corsOrigin *string
}

// serverCmd represents the 'server' command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the Sqedule API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbLogger, err := createLoggerWithLevel(*serverFlags.dbconn.dbLogLevel)
		if err != nil {
			return fmt.Errorf("Error initializing logger: %w", err)
		}

		db, err := dbutils.EstablishDatabaseConnection(
			*serverFlags.dbconn.dbType,
			*serverFlags.dbconn.dbConnection,
			&gorm.Config{
				Logger: dbLogger,
			})
		if err != nil {
			return fmt.Errorf("Error establishing database connection: %w", err)
		}

		engine := gin.Default()
		ctx := httpapi.Context{
			Db:         db,
			CorsOrigin: *serverFlags.corsOrigin,
		}

		err = ctx.SetupRouter(engine)
		if err != nil {
			return fmt.Errorf("Error setting up router: %w", err)
		}

		engine.Run(fmt.Sprintf("%s:%d", *serverFlags.bind, *serverFlags.port))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverFlags.dbconn = defineDatabaseConnectionFlags(serverCmd)

	serverFlags.bind = serverCmd.Flags().String("bind", serverDefaultBind, "IP to listen on")
	serverFlags.port = serverCmd.Flags().Int("port", serverDefaultPort, "port to listen on")
	serverFlags.corsOrigin = serverCmd.Flags().String("cors-origin", "", "CORS origin to allow")
}

package main

import (
	"context"
	"fmt"

	"github.com/fullstaq-labs/sqedule/cli"
	"github.com/fullstaq-labs/sqedule/server/approvalrulesprocessing"
	"github.com/fullstaq-labs/sqedule/server/dbmigrations"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/organizationmemberrole"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"github.com/fullstaq-labs/sqedule/server/httpapi"
	"github.com/fullstaq-labs/sqedule/server/webuiassetsserving"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
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

		err = runCmd_createDefaultOrg(viper.GetViper(), db, logger)
		if err != nil {
			return err
		}

		if !viper.GetBool("dev") {
			gin.SetMode(gin.ReleaseMode)
		}
		engine := gin.Default()
		ctx := httpapi.Context{
			Db:              db,
			DevelopmentMode: viper.GetBool("dev"),
			CorsOrigin:      viper.GetString("cors-origin"),
		}

		err = ctx.SetupRouter(engine, logger)
		if err != nil {
			return fmt.Errorf("Error setting up router: %w", err)
		}

		if webuiassetsserving.Enabled {
			err = webuiassetsserving.Intialize()
			if err != nil {
				return fmt.Errorf("Error initializing serving of web UI assets: %w", err)
			}
			err = webuiassetsserving.SetupRouter(engine)
			if err != nil {
				return fmt.Errorf("Error installing routes for serving web UI assets: %w", err)
			}
		}

		err = approvalrulesprocessing.ProcessAllPendingReleasesInBackground(ctx.Db)
		if err != nil {
			return fmt.Errorf("Error processing pending releases in the background: %w", err)
		}

		engine.Run(fmt.Sprintf("%s:%d", viper.GetString("bind"), viper.GetInt("port")))
		return nil
	},
}

func runCmd_createDefaultOrg(viper *viper.Viper, db *gorm.DB, logger gormlogger.Interface) error {
	// When removing this function, don't forget to also update the corresponding code in
	// - server/httpapi/auth/middleware_org_member_lookup.go, run()
	// - server/httpapi/router.go, installAuthenticationMiddlewares()
	var org dbmodels.Organization

	tx := db.Take(&org)
	err := dbutils.CreateFindOperationError(tx)
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("Error querying whether any Organizations are defined: %w", tx.Error)
	} else if err == nil {
		return nil
	}

	logger.Warn(context.Background(), "Creating default Organization")
	org = dbmodels.Organization{
		ID:          "default",
		DisplayName: "Default Organization",
	}
	if tx = db.Create(&org); tx.Error != nil {
		return fmt.Errorf("Error creating default Organization: %w", tx.Error)
	}

	user := dbmodels.User{
		OrganizationMember: dbmodels.OrganizationMember{
			BaseModel: dbmodels.BaseModel{
				OrganizationID: "default",
				Organization:   org,
			},
			Role:         organizationmemberrole.Admin,
			PasswordHash: "$argon2id$v=19$m=16,t=2,p=1$WlBFUmxyMkJWakw4TUMxVw$NyRkqa3o0uaAHnp7XpjU5A", // 123456
		},
		Email:     "nonexistant@default.org",
		FirstName: "Default",
		LastName:  "User",
	}
	if tx = db.Create(&user); tx.Error != nil {
		return fmt.Errorf("Error creating default user account: %w", tx.Error)
	}

	return nil
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

	flags.String("bind", "localhost", "IP to listen on")
	flags.Int("port", 3001, "port to listen on")
	flags.String("cors-origin", "", "CORS origin to allow")
	flags.Bool("auto-db-migrate", true, "automatically migrate database schema")
	flags.Bool("dev", false, "run in development mode")
}

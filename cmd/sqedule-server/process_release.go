package main

import (
	"fmt"

	"github.com/fullstaq-labs/sqedule/server/approvalrulesprocessing"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var processReleaseFlags struct {
	dbconn         databaseConnectionFlags
	organizationID *string
	applicationID  *string
	releaseID      *uint64
}

// processReleaseCmd represents the 'process-release' command
var processReleaseCmd = &cobra.Command{
	Use:   "process-release",
	Short: "Run the Sqedule API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbLogger, err := createLoggerWithLevel(*processReleaseFlags.dbconn.dbLogLevel)
		if err != nil {
			return fmt.Errorf("Error initializing logger: %w", err)
		}

		db, err := dbutils.EstablishDatabaseConnection(
			*processReleaseFlags.dbconn.dbType,
			*processReleaseFlags.dbconn.dbConnection,
			&gorm.Config{
				Logger: dbLogger,
			})
		if err != nil {
			return fmt.Errorf("Error establishing database connection: %w", err)
		}

		organization, err := dbmodels.FindOrganizationByID(db, *processReleaseFlags.organizationID)
		if err != nil {
			return fmt.Errorf("Error loading Organization: %w", err)
		}

		release, err := dbmodels.FindRelease(db, *processReleaseFlags.organizationID,
			*processReleaseFlags.applicationID, *processReleaseFlags.releaseID)
		if err != nil {
			return fmt.Errorf("Error loading Release: %w", err)
		}

		job, err := dbmodels.FindReleaseBackgroundJob(db.Preload("Release"),
			organization.ID, *processReleaseFlags.applicationID, release.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				job, err = dbmodels.CreateReleaseBackgroundJob(db, organization.ID,
					*processReleaseFlags.applicationID, release)
				if err != nil {
					return fmt.Errorf("Error creating ReleaseBackgroundJob: %w", err)
				}
			} else {
				return fmt.Errorf("Error finding ReleaseBackgroundJob: %w", err)
			}
		}

		engine := approvalrulesprocessing.Engine{
			Db:                   db,
			Organization:         organization,
			ReleaseBackgroundJob: job,
		}
		return engine.Run()
	},
}

func init() {
	cmd := processReleaseCmd
	flags := cmd.Flags()
	rootCmd.AddCommand(cmd)

	processReleaseFlags.dbconn = defineDatabaseConnectionFlags(cmd)

	processReleaseFlags.organizationID = flags.String("organization-id", "", "")
	processReleaseFlags.applicationID = flags.String("application-id", "", "")
	processReleaseFlags.releaseID = flags.Uint64("release-id", 0, "")
}

package approvalrulesprocessing

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/lib/mocking"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"gorm.io/gorm"
)

const (
	backgroundProcessingRetryMinDuration = 5 * time.Second
	backgroundProcessingRetryMaxDuration = 5 * time.Minute
	backgroundProcessingRetryJitter      = 10 * time.Second
	backgroundProcessingRetryMaxAttempts = 10
)

func ProcessInBackground(db *gorm.DB, organizationID string, job dbmodels.ReleaseBackgroundJob) error {
	go realProcessInBackground(db, organizationID, job, mocking.RealClock{}, false)
	return nil
}

func realProcessInBackground(db *gorm.DB, organizationID string, job dbmodels.ReleaseBackgroundJob, clock mocking.IClock, fakeError bool) error {
	var lastSleepDuration time.Duration
	var retryCount uint
	var err error

	for {
		engine := Engine{Db: db, OrganizationID: organizationID, ReleaseBackgroundJob: job}
		if fakeError {
			err = errors.New("fake error")
		} else {
			err = engine.Run()
		}
		if err == nil {
			return nil
		}

		if retryCount == backgroundProcessingRetryMaxAttempts {
			db.Logger.Error(context.Background(), "Error processing release %s. Already retried %d times, so will no longer retry. Error: %s",
				job.Release.Description(), retryCount, err.Error())
			return err
		}

		retryCount++
		lastSleepDuration = calcRetrySleepDuration(lastSleepDuration)
		db.Logger.Error(context.Background(), "Error processing release %s. Will retry (attempt %d/%d) in %.0f seconds. Error: %s",
			job.Release.Description(), retryCount, backgroundProcessingRetryMaxAttempts, lastSleepDuration.Seconds(), err.Error())
		clock.Sleep(lastSleepDuration)
	}
}

func calcRetrySleepDuration(lastDuration time.Duration) time.Duration {
	jitter := rand.Int63n(int64(backgroundProcessingRetryJitter)) - (int64(backgroundProcessingRetryJitter) / 2)
	return time.Duration(lib.AddAndClamp(uint64(lastDuration*2), jitter,
		uint64(backgroundProcessingRetryMinDuration), uint64(backgroundProcessingRetryMaxDuration)))
}

func ProcessAllPendingReleasesInBackground(db *gorm.DB) error {
	jobs, err := dbmodels.FindUnlockedReleaseBackgroundJobs(db)
	if err != nil {
		return fmt.Errorf("Error querying release background jobs: %w", err)
	}

	for _, job := range jobs {
		err = ProcessInBackground(db, job.OrganizationID, job)
		if err != nil {
			return fmt.Errorf("Error processing release %s in background: %w", job.Release.Description(), err)
		}
	}

	return nil
}

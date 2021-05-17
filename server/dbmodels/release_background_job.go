package dbmodels

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReleaseBackgroundJobPostgresLockNamespace is the number at which the PostgreSQL advisory lock should start.
// Given a ReleaseBackgroundJob with a certain LockID, the corresponding advisory lock ID is
// `ReleaseBackgroundJobPostgresLockNamespace + LockID`
const ReleaseBackgroundJobPostgresLockNamespace uint64 = 1 * 0xffffffff

// ReleaseBackgroundJobMaxLockID is the maximum value that `ReleaseBackgroundJob.LockID` may have.
var ReleaseBackgroundJobMaxLockSubID uint32 = uint32(math.Pow(2, 31)) - 1

// ReleaseBackgroundJob ...
type ReleaseBackgroundJob struct {
	BaseModel
	ApplicationID string    `gorm:"type:citext; primaryKey; not null"`
	ReleaseID     uint64    `gorm:"primaryKey; not null"`
	Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LockSubID     uint32    `gorm:"type:int; autoIncrement; unique; not null; check:(lock_sub_id > 0)"`
	CreatedAt     time.Time `gorm:"not null"`
}

// CreateReleaseBackgroundJob ...
func CreateReleaseBackgroundJob(db *gorm.DB, organizationID string, applicationID string,
	release Release) (ReleaseBackgroundJob, error) {
	job, _, err := createReleaseBackgroundJobWithDebug(db, organizationID, applicationID, release, 1000)
	return job, err
}

func createReleaseBackgroundJobWithDebug(db *gorm.DB, organizationID string, applicationID string,
	release Release, maxTries uint) (ReleaseBackgroundJob, uint, error) {
	var job ReleaseBackgroundJob
	var numTry uint = 0
	var created bool = false
	var err error

	// Keep trying to create a job (and its related job ruleset bindings)
	// until we've successfully picked a unique lock ID or encountered an error.

	for ; numTry < maxTries && !created && err == nil; numTry++ {
		err = db.Transaction(func(tx *gorm.DB) error {
			job = ReleaseBackgroundJob{
				BaseModel: BaseModel{
					OrganizationID: organizationID,
				},
				ApplicationID: applicationID,
				ReleaseID:     release.ID,
				Release:       release,
			}
			if numTry > 0 {
				// We were unable to obtain a free lock sub-ID through auto-incrementation.
				// So pick a random one instead.
				job.LockSubID = uint32(uint64(rand.Uint32()) % (uint64(ReleaseBackgroundJobMaxLockSubID) + 1))
			}
			savetx := tx.Omit(clause.Associations).Create(&job)
			if savetx.Error != nil {
				return savetx.Error
			}

			return nil
		})

		if err != nil {
			if dbutils.IsUniqueConstraintError(err, "release_background_jobs_lock_sub_id_key") {
				// Try again
				err = nil
			}
		} else {
			created = true
		}
	}

	if created {
		return job, numTry, nil
	}
	if err != nil {
		return ReleaseBackgroundJob{}, numTry, err
	}
	return ReleaseBackgroundJob{}, numTry, fmt.Errorf("Unable to find a free lock sub-ID after %d tries", maxTries)
}

// FindReleaseBackgroundJob looks up a ReleaseBackgroundJob by its application ID and release ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindReleaseBackgroundJob(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) (ReleaseBackgroundJob, error) {
	var result ReleaseBackgroundJob

	tx := db.Where("organization_id = ? AND application_id = ? AND release_id = ?", organizationID, applicationID, releaseID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

// FindUnlockedReleaseBackgroundJobs returns all ReleaseBackgroundJobs, in the entire database
// (across organizations), that aren't currently being processed by approvalrulesprocessing.Engine.
func FindUnlockedReleaseBackgroundJobs(db *gorm.DB) ([]ReleaseBackgroundJob, error) {
	var result []ReleaseBackgroundJob
	tx := db.Clauses(clause.Locking{Strength: "UPDATE SKIP LOCKED"}).Preload("Release").Find(&result)
	return result, tx.Error
}

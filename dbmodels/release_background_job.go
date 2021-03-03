package dbmodels

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReleaseBackgroundJobPostgresLockNamespace is the number at which the PostgreSQL advisory lock should start.
// Given a ReleaseBackgroundJob with a certain LockID, the corresponding advisory lock ID is
// `ReleaseBackgroundJobPostgresLockNamespace + LockID`
const ReleaseBackgroundJobPostgresLockNamespace uint64 = 1 * 0xffffffff

// ReleaseBackgroundJobMaxLockID is the maximum value that `ReleaseBackgroundJob.LockID` may have.
var ReleaseBackgroundJobMaxLockID uint32 = uint32(math.Pow(2, 31)) - 1

// ReleaseBackgroundJob ...
type ReleaseBackgroundJob struct {
	BaseModel
	ApplicationID string    `gorm:"type:citext; primaryKey; not null"`
	ReleaseID     uint64    `gorm:"primaryKey; not null"`
	Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LockID        uint32    `gorm:"type:int; autoIncrement; unique; not null; check:(lock_id > 0)"`
	CreatedAt     time.Time `gorm:"not null"`
}

// ReleaseBackgroundJobApprovalRulesetBinding ...
type ReleaseBackgroundJobApprovalRulesetBinding struct {
	BaseModel

	ApplicationID        string               `gorm:"type:citext; primaryKey; not null"`
	ReleaseID            uint64               `gorm:"primaryKey; not null"`
	ReleaseBackgroundJob ReleaseBackgroundJob `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ReleaseID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
	ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetMajorVersionID uint64                      `gorm:"not null"`
	ApprovalRulesetMajorVersion   ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetMinorVersionNumber uint32                      `gorm:"type:int; not null"`
	ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
}

// CreateReleaseBackgroundJob ...
func CreateReleaseBackgroundJob(db *gorm.DB, organization Organization, applicationID string,
	release Release) (ReleaseBackgroundJob, error) {
	job, _, err := createReleaseBackgroundJobWithDebug(db, organization, applicationID, release, 1000)
	return job, err
}

func createReleaseBackgroundJobWithDebug(db *gorm.DB, organization Organization, applicationID string,
	release Release, maxTries uint) (ReleaseBackgroundJob, uint, error) {
	var job ReleaseBackgroundJob
	var numTry uint = 0
	var created bool = false
	var err error

	// Load related ruleset bindings

	bindings, err := FindAllApprovalRulesetBindings(db.Preload("ApprovalRuleset"),
		organization.ID, applicationID)
	if err != nil {
		return ReleaseBackgroundJob{}, 0, err
	}
	err = LoadApprovalRulesetsLatestVersions(db, organization.ID,
		CollectApprovalRulesetBindingRulesets(bindings))
	if err != nil {
		return ReleaseBackgroundJob{}, 0, err
	}

	// Keep trying to create a job (and its related job ruleset bindings)
	// until we've successfully picked a unique lock ID or encountered an error.

	for ; numTry < maxTries && !created && err == nil; numTry++ {
		err = db.Transaction(func(tx *gorm.DB) error {
			job = ReleaseBackgroundJob{
				BaseModel: BaseModel{
					OrganizationID: organization.ID,
					Organization:   organization,
				},
				ApplicationID: applicationID,
				ReleaseID:     release.ID,
				Release:       release,
			}
			if numTry > 0 {
				// We were unable to obtain a free lock ID through auto-incrementation.
				// So pick a random one instead.
				job.LockID = rand.Uint32() % ReleaseBackgroundJobMaxLockID
			}
			savetx := tx.Omit(clause.Associations).Create(&job)
			if savetx.Error != nil {
				return savetx.Error
			}

			// Create release job ruleset bindings
			for _, binding := range bindings {
				jobBinding := ReleaseBackgroundJobApprovalRulesetBinding{
					BaseModel: BaseModel{
						OrganizationID: organization.ID,
					},
					ApplicationID:                     applicationID,
					ReleaseID:                         release.ID,
					ApprovalRulesetID:                 binding.ApprovalRulesetID,
					ApprovalRulesetMajorVersionID:     binding.ApprovalRuleset.LatestMajorVersion.ID,
					ApprovalRulesetMinorVersionNumber: binding.ApprovalRuleset.LatestMinorVersion.VersionNumber,
					Mode:                              binding.Mode,
				}
				savetx = tx.Create(&jobBinding)
				if savetx.Error != nil {
					return savetx.Error
				}
			}

			return nil
		})

		if err != nil {
			if dbutils.IsUniqueConstraintError(err, "release_background_jobs_lock_id_key") {
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
	return ReleaseBackgroundJob{}, numTry, fmt.Errorf("Unable to find a free lock ID after %d tries", maxTries)
}

// FindReleaseBackgroundJob looks up a ReleaseBackgroundJob by its application ID and release ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindReleaseBackgroundJob(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) (ReleaseBackgroundJob, error) {
	var result ReleaseBackgroundJob

	tx := db.Where("organization_id = ? AND application_id = ? AND release_id = ?", organizationID, applicationID, releaseID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

// FindAllReleaseBackgroundJobApprovalRulesetBindings ...
func FindAllReleaseBackgroundJobApprovalRulesetBindings(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) ([]ReleaseBackgroundJobApprovalRulesetBinding, error) {
	var result []ReleaseBackgroundJobApprovalRulesetBinding
	tx := db.Where("organization_id = ? AND application_id = ? AND release_id = ?",
		organizationID, applicationID, releaseID)
	tx = tx.Find(&result)
	return result, tx.Error
}

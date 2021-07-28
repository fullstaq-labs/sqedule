package dbmodels

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

type Release struct {
	BaseModel
	ApplicationID  string             `gorm:"type:citext; primaryKey; not null"`
	Application    Application        `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ID             uint64             `gorm:"primaryKey; not null"`
	State          releasestate.State `gorm:"type:release_state; not null"`
	SourceIdentity sql.NullString
	Comments       sql.NullString
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
	FinalizedAt    sql.NullTime
}

//
// ******** Release methods ********
//

func (r Release) Description() string {
	return fmt.Sprintf("(org=%s, app=%s, releaseID=%d)", r.OrganizationID, r.ApplicationID, r.ID)
}

//
// ******** Find/load functions ********
//

func FindReleases(db *gorm.DB, organizationID string, applicationID string) ([]Release, error) {
	var result []Release
	tx := db.Where("organization_id = ?", organizationID)
	if len(applicationID) > 0 {
		tx = tx.Where("application_id = ?", applicationID)
	}
	tx = tx.Find(&result)
	return result, tx.Error
}

// FindRelease looks up a Release by its ID and its application ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindRelease(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) (Release, error) {
	var result Release

	tx := db.Where("organization_id = ? AND application_id = ? AND id = ?", organizationID, applicationID, releaseID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

//
// ******** Other functions ********
//

func MakeReleasesPointerArray(releases []Release) []*Release {
	result := make([]*Release, 0, len(releases))
	for i := range releases {
		result = append(result, &releases[i])
	}
	return result
}

func CollectReleasesWithReleaseApprovalRulesetBindings(bindings []ReleaseApprovalRulesetBinding) []*Release {
	result := make([]*Release, 0, len(bindings))
	for i := range bindings {
		binding := &bindings[i]
		result = append(result, &binding.Release)
	}
	return result
}

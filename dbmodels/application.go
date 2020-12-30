package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"gorm.io/gorm"
)

// Application ...
type Application struct {
	BaseModel
	ID                 string                   `gorm:"type:citext; primaryKey; not null"`
	CreatedAt          time.Time                `gorm:"not null"`
	LatestMajorVersion *ApplicationMajorVersion `gorm:"-"`
	LatestMinorVersion *ApplicationMinorVersion `gorm:"-"`
}

// ApplicationMajorVersion ...
type ApplicationMajorVersion struct {
	OrganizationID string       `gorm:"type:citext; primaryKey; not null; index:application_major_version_idx,unique"`
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
	ApplicationID  string       `gorm:"type:citext; not null; index:application_major_version_idx,unique"`
	VersionNumber  *uint32      `gorm:"index:application_major_version_idx,sort:desc,where:version_number IS NOT NULL,unique"`
	CreatedAt      time.Time    `gorm:"not null"`
	UpdatedAt      time.Time    `gorm:"not null"`

	Application Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApplicationMinorVersion ...
type ApplicationMinorVersion struct {
	BaseModel
	ApplicationMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber             uint32            `gorm:"primaryKey; not null"`
	ReviewState               reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments            sql.NullString
	CreatedAt                 time.Time `gorm:"not null"`
	Enabled                   bool      `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`

	ApplicationMajorVersion ApplicationMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// FindApplication looks up an Application by its ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindApplication(db *gorm.DB, organizationID string, id string) (Application, error) {
	var result Application

	tx := db.Where("organization_id = ? AND id = ?", organizationID, id)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

// LoadApplicationsLatestVersions ...
func LoadApplicationsLatestVersions(db *gorm.DB, organizationID string, applications []*Application) error {
	if len(applications) == 0 {
		return nil
	}

	appIndex := make(map[string][]*Application)
	appIds := make([]string, 0, len(applications))
	for _, app := range applications {
		if appIndex[app.ID] == nil {
			appIndex[app.ID] = make([]*Application, 0)
			appIds = append(appIds, app.ID)
		}
		appIndex[app.ID] = append(appIndex[app.ID], app)
	}

	majorVersions := make([]ApplicationMajorVersion, 0)
	majorIndex := make(map[uint64]*ApplicationMajorVersion)
	majorIds := make([]uint64, 0)
	tx := db.
		Select("DISTINCT ON (organization_id, application_id) *").
		Where("organization_id = ? AND application_id IN ? AND version_number IS NOT NULL",
			organizationID, appIds).
		Order("organization_id, application_id, version_number DESC").
		Find(&majorVersions)
	if tx.Error != nil {
		return tx.Error
	}
	for _, majorVersion := range majorVersions {
		apps := appIndex[majorVersion.ApplicationID]
		for _, app := range apps {
			app.LatestMajorVersion = &majorVersion
		}
		majorVersion.Application = *apps[0]
		majorIndex[majorVersion.ID] = &majorVersion
		majorIds = append(majorIds, majorVersion.ID)
	}

	var minorVersions []ApplicationMinorVersion
	tx = db.
		Select("DISTINCT ON (organization_id, application_major_version_id) *").
		Where("organization_id = ? AND application_major_version_id IN ?",
			organizationID, majorIds).
		Order("organization_id, application_major_version_id, version_number DESC").
		Find(&minorVersions)
	if tx.Error != nil {
		return tx.Error
	}
	for _, minorVersion := range minorVersions {
		majorVersion := majorIndex[minorVersion.ApplicationMajorVersionID]
		minorVersion.ApplicationMajorVersion = *majorVersion
		apps := appIndex[majorVersion.ApplicationID]
		for _, app := range apps {
			app.LatestMinorVersion = &minorVersion
		}
	}

	return nil
}

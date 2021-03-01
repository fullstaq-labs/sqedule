package dbmodels

import (
	"database/sql"
	"reflect"
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
	VersionNumber  *uint32      `gorm:"type:int; index:application_major_version_idx,sort:desc,where:version_number IS NOT NULL,unique; check:(version_number > 0)"`
	CreatedAt      time.Time    `gorm:"not null"`
	UpdatedAt      time.Time    `gorm:"not null"`

	Application Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApplicationMinorVersion ...
type ApplicationMinorVersion struct {
	BaseModel
	ApplicationMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber             uint32            `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
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

// GetID ...
func (app Application) GetID() interface{} {
	return app.ID
}

// SetLatestMajorVersion ...
func (app *Application) SetLatestMajorVersion(majorVersion IReviewableMajorVersion) {
	app.LatestMajorVersion = majorVersion.(*ApplicationMajorVersion)
}

// SetLatestMinorVersion ...
func (app *Application) SetLatestMinorVersion(minorVersion IReviewableMinorVersion) {
	app.LatestMinorVersion = minorVersion.(*ApplicationMinorVersion)
}

// GetID ...
func (major ApplicationMajorVersion) GetID() interface{} {
	return major.ID
}

// GetReviewableID ...
func (major ApplicationMajorVersion) GetReviewableID() interface{} {
	return major.ApplicationID
}

// AssociateWithReviewable ...
func (major *ApplicationMajorVersion) AssociateWithReviewable(reviewable IReviewable) {
	application := reviewable.(*Application)
	major.ApplicationID = application.ID
	major.Application = *application
}

// GetMajorVersionID ...
func (minor ApplicationMinorVersion) GetMajorVersionID() interface{} {
	return minor.ApplicationMajorVersionID
}

// AssociateWithMajorVersion ...
func (minor *ApplicationMinorVersion) AssociateWithMajorVersion(majorVersion IReviewableMajorVersion) {
	concreteMajorVersion := majorVersion.(*ApplicationMajorVersion)
	minor.ApplicationMajorVersionID = concreteMajorVersion.ID
	minor.ApplicationMajorVersion = *concreteMajorVersion
}

// LoadApplicationsLatestVersions ...
func LoadApplicationsLatestVersions(db *gorm.DB, organizationID string, applications []*Application) error {
	reviewables := make([]IReviewable, 0, len(applications))
	for _, app := range applications {
		reviewables = append(reviewables, app)
	}

	return LoadReviewablesLatestVersions(
		db,
		reflect.TypeOf(""),
		"application_id",
		reflect.TypeOf(ApplicationMajorVersion{}),
		reflect.TypeOf(uint64(0)),
		"application_major_version_id",
		reflect.TypeOf(ApplicationMinorVersion{}),
		organizationID,
		reviewables,
	)
}

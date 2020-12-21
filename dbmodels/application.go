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
	ID        string    `gorm:"type:citext; primaryKey; not null"`
	CreatedAt time.Time `gorm:"not null"`
}

// ApplicationMajorVersion ...
type ApplicationMajorVersion struct {
	OrganizationID string       `gorm:"type:citext; primaryKey; not null; index:application_major_version_idx,unique"`
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
	ApplicationID  string       `gorm:"type:citext; not null; index:application_major_version_idx,unique"`
	VersionNumber  *uint32      `gorm:"index:application_major_version_idx,unique"`
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

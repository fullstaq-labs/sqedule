package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/reviewstate"
)

type Application struct {
	BaseModel
	ID        string    `gorm:"primaryKey; not null"`
	CreatedAt time.Time `gorm:"not null"`
}

type ApplicationMajorVersion struct {
	OrganizationID string    `gorm:"primaryKey; not null; index:version,unique"`
	ID             uint64    `gorm:"primaryKey; not null"`
	ApplicationID  string    `gorm:"index:version,unique"`
	VersionNumber  *uint32   `gorm:"index:version,unique"`
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`

	Application Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type ApplicationMinorVersion struct {
	BaseModel
	ApplicationMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber             uint32            `gorm:"primaryKey; not null"`
	ReviewState               reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments            sql.NullString
	CreatedAt                 time.Time `gorm:"not null"`
	Enabled                   bool      `gorm:"not null"`

	DisplayName string `gorm:"not null"`

	ApplicationMajorVersion ApplicationMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

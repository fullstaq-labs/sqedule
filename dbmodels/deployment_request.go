package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
)

type DeploymentRequest struct {
	BaseModel
	ID             uint64                       `gorm:"primaryKey; not null"`
	State          deploymentrequeststate.State `gorm:"type:deployment_request_state; not null"`
	SourceIdentity sql.NullString
	Comments       sql.NullString
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
	FinalizedAt    sql.NullTime

	ApplicationMajorVersionID     uint64                  `gorm:"not null"`
	ApplicationMinorVersionNumber uint32                  `gorm:"not null"`
	ApplicationMinorVersion       ApplicationMinorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID,ApplicationMinorVersionNumber; references:OrganizationID,ApplicationMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

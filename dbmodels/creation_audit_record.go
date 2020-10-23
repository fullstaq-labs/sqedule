package dbmodels

import (
	"database/sql"
	"time"
)

type CreationAuditRecord struct {
	BaseModel
	ID                   uint64 `gorm:"primaryKey; not null"`
	OrganizationMemberIP sql.NullString
	CreatedAt            time.Time `gorm:"not null"`

	// Object association

	UserEmail sql.NullString `gorm:"type: citext"`
	User      User           `gorm:"foreignKey:OrganizationID,UserEmail; references:OrganizationID,Email; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT; check:((CASE user_email IS NULL THEN 0 ELSE 1 END) + (CASE service_account_name IS NULL THEN 0 ELSE 1 END) <= 1)"`

	ServiceAccountName sql.NullString `gorm:"type: citext"`
	ServiceAccount     ServiceAccount `gorm:"foreignKey:OrganizationID,ServiceAccountName; references:OrganizationID,Name; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	// Subject association

	ApplicationMajorVersionID     uint64
	ApplicationMinorVersionNumber uint32                  `gorm:"check:((application_major_version_id IS NULL) = (application_minor_version_number IS NULL))"`
	ApplicationMinorVersion       ApplicationMinorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID,ApplicationMinorVersionNumber; references:OrganizationID,ApplicationMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT; check:((CASE application_minor_version_number IS NULL THEN 0 ELSE 1 END) + (CASE deployment_request_id IS NULL THEN 0 ELSE 1 END) <= 1)"`

	DeploymentRequestID uint64
	DeploymentRequest   DeploymentRequest `gorm:"foreignKey:OrganizationID,DeploymentRequestID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

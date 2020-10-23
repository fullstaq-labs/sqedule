package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000070)
}

var migration20201021000070 = gormigrate.Migration{
	ID: "20201021000070 Creation audit record",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type: citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type: citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type OrganizationMember struct {
			BaseModel
		}

		type User struct {
			OrganizationMember
			Email string `gorm:"type: citext; primaryKey; not null"`
		}

		type ServiceAccount struct {
			OrganizationMember
			Name string `gorm:"type: citext; primaryKey; not null"`
		}

		type Application struct {
			BaseModel
			ID string `gorm:"type: citext; primaryKey; not null"`
		}

		type ApplicationMajorVersion struct {
			OrganizationID string       `gorm:"type: citext; primaryKey; not null; index:version,unique"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
			ApplicationID  string       `gorm:"type: citext; not null; index:version,unique"`
			Application    Application  `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ApplicationMinorVersion struct {
			BaseModel
			ApplicationMajorVersionID uint64                  `gorm:"primaryKey; not null"`
			VersionNumber             uint32                  `gorm:"primaryKey; not null"`
			ApplicationMajorVersion   ApplicationMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type DeploymentRequest struct {
			BaseModel
			ID uint64 `gorm:"primaryKey; not null"`

			ApplicationMajorVersionID     uint64                  `gorm:"not null"`
			ApplicationMinorVersionNumber uint32                  `gorm:"not null"`
			ApplicationMinorVersion       ApplicationMinorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID,ApplicationMinorVersionNumber; references:OrganizationID,ApplicationMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

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

		return tx.AutoMigrate(&CreationAuditRecord{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("creation_audit_records")
	},
}

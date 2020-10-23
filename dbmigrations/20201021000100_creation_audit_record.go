package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000100)
}

var migration20201021000100 = gormigrate.Migration{
	ID: "20201021000100 Creation audit record",
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

		type ApplicationMinorVersion struct {
			BaseModel
			ApplicationMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber             uint32 `gorm:"primaryKey; not null"`
		}

		type ApprovalRulesetMinorVersion struct {
			BaseModel
			ApprovalRulesetMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber                 uint32 `gorm:"primaryKey; not null"`
		}

		type DeploymentRequestEvent struct {
			BaseModel
			ID uint64 `gorm:"primaryKey; not null"`
		}

		type DeploymentRequestCreatedEvent struct {
			DeploymentRequestEvent
		}

		type DeploymentRequestCancelledEvent struct {
			DeploymentRequestEvent
		}

		type DeploymentRequestRuleProcessedEvent struct {
			DeploymentRequestEvent
		}

		type CreationAuditRecord struct {
			BaseModel
			ID                   uint64 `gorm:"primaryKey; not null"`
			OrganizationMemberIP sql.NullString
			CreatedAt            time.Time `gorm:"not null"`

			// Object association

			UserEmail sql.NullString `gorm:"type: citext; check:((CASE WHEN user_email IS NULL THEN 0 ELSE 1 END) + (CASE WHEN service_account_name IS NULL THEN 0 ELSE 1 END) <= 1)"`
			User      User           `gorm:"foreignKey:OrganizationID,UserEmail; references:OrganizationID,Email; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ServiceAccountName sql.NullString `gorm:"type: citext"`
			ServiceAccount     ServiceAccount `gorm:"foreignKey:OrganizationID,ServiceAccountName; references:OrganizationID,Name; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			// Subject association

			ApplicationMajorVersionID     uint64                  `gorm:"check:((CASE WHEN application_minor_version_number IS NULL THEN 0 ELSE 1 END) + (CASE WHEN approval_ruleset_minor_version_number IS NULL THEN 0 ELSE 1 END) + (CASE WHEN deployment_request_created_event_id IS NULL THEN 0 ELSE 1 END) + (CASE WHEN deployment_request_cancelled_event_id IS NULL THEN 0 ELSE 1 END) = 1)"`
			ApplicationMinorVersionNumber uint32                  `gorm:"check:((application_major_version_id IS NULL) = (application_minor_version_number IS NULL))"`
			ApplicationMinorVersion       ApplicationMinorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID,ApplicationMinorVersionNumber; references:OrganizationID,ApplicationMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ApprovalRulesetMajorVersionID     uint64
			ApprovalRulesetMinorVersionNumber uint32                      `gorm:"check:((approval_ruleset_major_version_id IS NULL) = (approval_ruleset_minor_version_number IS NULL))"`
			ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber"`

			DeploymentRequestCreatedEventID uint64
			DeploymentRequestCreatedEvent   DeploymentRequestCreatedEvent `gorm:"foreignKey:OrganizationID,DeploymentRequestCreatedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			DeploymentRequestCancelledEventID uint64
			DeploymentRequestCancelledEvent   DeploymentRequestCancelledEvent `gorm:"foreignKey:OrganizationID,DeploymentRequestCancelledEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		return tx.AutoMigrate(&CreationAuditRecord{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("creation_audit_records")
	},
}
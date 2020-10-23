package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000080)
}

var migration20201021000080 = gormigrate.Migration{
	ID: "20201021000080 Deployment request event",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type Application struct {
			BaseModel
			ID string `gorm:"primaryKey; not null"`
		}

		type ApplicationMajorVersion struct {
			OrganizationID string       `gorm:"primaryKey; not null; index:version,unique"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
			ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
			ApplicationID  string       `gorm:"not null; index:version,unique"`
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

		type DeploymentRequestEvent struct {
			BaseModel
			ID                  uint64            `gorm:"primaryKey; not null"`
			DeploymentRequestID uint64            `gorm:"not null"`
			DeploymentRequest   DeploymentRequest `gorm:"foreignKey:OrganizationID,DeploymentRequestID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
			CreatedAt           time.Time         `gorm:"not null"`
		}

		type DeploymentRequestCreatedEvent struct {
			DeploymentRequestEvent
		}

		type DeploymentRequestCancelledEvent struct {
			DeploymentRequestEvent
		}

		type DeploymentRequestRuleProcessedEvent struct {
			DeploymentRequestEvent
			ResultState string `gorm:"type:deployment_request_state; not null"`
		}

		return tx.AutoMigrate(&DeploymentRequestEvent{}, &DeploymentRequestCreatedEvent{},
			&DeploymentRequestCancelledEvent{}, &DeploymentRequestRuleProcessedEvent{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("deployment_request_rule_processed_events",
			"deployment_request_cancelled_events", "deployment_request_created_events",
			"deployment_request_events")
	},
}

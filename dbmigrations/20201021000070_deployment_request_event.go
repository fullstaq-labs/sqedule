package dbmigrations

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000070)
}

var migration20201021000070 = gormigrate.Migration{
	ID: "20201021000070 Deployment request event",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type: citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type: citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type DeploymentRequest struct {
			BaseModel
			ID uint64 `gorm:"primaryKey; not null"`
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

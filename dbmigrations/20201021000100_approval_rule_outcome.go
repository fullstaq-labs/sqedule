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
	ID: "20201021000100 Approval rule outcome",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
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
		}

		type DeploymentRequestRuleProcessedEvent struct {
			DeploymentRequestEvent
		}

		// ApprovalRuleOutcome ...
		type ApprovalRuleOutcome struct {
			BaseModel
			ID                                    uint64                              `gorm:"primaryKey; autoIncrement; not null"`
			DeploymentRequestRuleProcessedEventID uint64                              `gorm:"not null"`
			DeploymentRequestRuleProcessedEvent   DeploymentRequestRuleProcessedEvent `gorm:"foreignKey:OrganizationID,DeploymentRequestRuleProcessedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
			CreatedAt                             time.Time                           `gorm:"not null"`
		}

		// HTTPApiApprovalRuleOutcome ...
		type HTTPApiApprovalRuleOutcome struct {
			ApprovalRuleOutcome
			ResponseCode        uint8  `gorm:"not null"`
			ResponseContentType string `gorm:"not null"`
			ResponseBody        []byte `gorm:"not null"`
		}

		// ScheduleApprovalRuleOutcome ...
		type ScheduleApprovalRuleOutcome struct {
			ApprovalRuleOutcome
		}

		// ManualApprovalRuleOutcome ...
		type ManualApprovalRuleOutcome struct {
			ApprovalRuleOutcome
			Comments sql.NullString
		}

		return tx.AutoMigrate(&HTTPApiApprovalRuleOutcome{}, &ScheduleApprovalRuleOutcome{},
			&ManualApprovalRuleOutcome{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("http_api_approval_rule_outcomes",
			"schedule_approval_rule_outcomes", "manual_approval_rule_outcomes")
	},
}

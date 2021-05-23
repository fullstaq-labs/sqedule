package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000150)
}

var migration20201021000150 = gormigrate.Migration{
	ID: "20201021000150 Creation audit record",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type OrganizationMember struct {
			BaseModel
		}

		type User struct {
			OrganizationMember
			Email string `gorm:"type:citext; primaryKey; not null"`
		}

		type ServiceAccount struct {
			OrganizationMember
			Name string `gorm:"type:citext; primaryKey; not null"`
		}

		type ReviewableAdjustmentBase struct {
			AdjustmentNumber uint32 `gorm:"type:int; primaryKey; not null; check:(adjustment_number > 0)"`
		}

		type ApplicationAdjustment struct {
			BaseModel
			ApplicationVersionID uint64 `gorm:"primaryKey; not null"`
			ReviewableAdjustmentBase
		}

		type ApprovalRulesetAdjustment struct {
			BaseModel
			ApprovalRulesetVersionID uint64 `gorm:"primaryKey; not null"`
			ReviewableAdjustmentBase
		}

		type ReleaseEvent struct {
			BaseModel
			ID uint64 `gorm:"primaryKey; not null"`
		}

		type ReleaseCreatedEvent struct {
			ReleaseEvent
		}

		type ReleaseCancelledEvent struct {
			ReleaseEvent
		}

		type ReleaseRuleProcessedEvent struct {
			ReleaseEvent
		}

		type ApprovalRuleOutcome struct {
			BaseModel
			ID                          uint64                    `gorm:"primaryKey; autoIncrement; not null"`
			ReleaseRuleProcessedEventID uint64                    `gorm:"not null"`
			ReleaseRuleProcessedEvent   ReleaseRuleProcessedEvent `gorm:"foreignKey:OrganizationID,ReleaseRuleProcessedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		type ManualApprovalRuleOutcome struct {
			ApprovalRuleOutcome
			Comments sql.NullString
		}

		type CreationAuditRecord struct {
			BaseModel
			ID                   uint64 `gorm:"primaryKey; not null"`
			OrganizationMemberIP sql.NullString
			CreatedAt            time.Time `gorm:"not null"`

			// Object association

			UserEmail sql.NullString `gorm:"type:citext; check:((CASE WHEN user_email IS NULL THEN 0 ELSE 1 END) + (CASE WHEN service_account_name IS NULL THEN 0 ELSE 1 END) <= 1)"`
			User      User           `gorm:"foreignKey:OrganizationID,UserEmail; references:OrganizationID,Email; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ServiceAccountName sql.NullString `gorm:"type:citext"`
			ServiceAccount     ServiceAccount `gorm:"foreignKey:OrganizationID,ServiceAccountName; references:OrganizationID,Name; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			// Subject association

			ApplicationVersionID        *uint64               `gorm:"check:((CASE WHEN application_adjustment_number IS NULL THEN 0 ELSE 1 END) + (CASE WHEN approval_ruleset_adjustment_number IS NULL THEN 0 ELSE 1 END) + (CASE WHEN manual_approval_rule_outcome_id IS NULL THEN 0 ELSE 1 END) + (CASE WHEN release_created_event_id IS NULL THEN 0 ELSE 1 END) + (CASE WHEN release_cancelled_event_id IS NULL THEN 0 ELSE 1 END) = 1)"`
			ApplicationAdjustmentNumber *uint32               `gorm:"type:int; check:((application_version_id IS NULL) = (application_adjustment_number IS NULL))"`
			ApplicationAdjustment       ApplicationAdjustment `gorm:"foreignKey:OrganizationID,ApplicationVersionID,ApplicationAdjustmentNumber; references:OrganizationID,ApplicationVersionID,AdjustmentNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ApprovalRulesetVersionID        *uint64
			ApprovalRulesetAdjustmentNumber *uint32                   `gorm:"type:int; check:((approval_ruleset_version_id IS NULL) = (approval_ruleset_adjustment_number IS NULL))"`
			ApprovalRulesetAdjustment       ApprovalRulesetAdjustment `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID,ApprovalRulesetAdjustmentNumber; references:OrganizationID,ApprovalRulesetVersionID,AdjustmentNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ManualApprovalRuleOutcomeID *uint64
			ManualApprovalRuleOutcome   ManualApprovalRuleOutcome `gorm:"foreignKey:OrganizationID,ManualApprovalRuleOutcomeID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ReleaseCreatedEventID *uint64
			ReleaseCreatedEvent   ReleaseCreatedEvent `gorm:"foreignKey:OrganizationID,ReleaseCreatedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

			ReleaseCancelledEventID *uint64
			ReleaseCancelledEvent   ReleaseCancelledEvent `gorm:"foreignKey:OrganizationID,ReleaseCancelledEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
		}

		return tx.AutoMigrate(&CreationAuditRecord{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable("creation_audit_records")
	},
}

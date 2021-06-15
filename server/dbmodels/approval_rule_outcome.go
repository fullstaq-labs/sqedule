package dbmodels

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********/
//

type ApprovalRuleOutcome struct {
	BaseModel
	ID                          uint64                    `gorm:"primaryKey; autoIncrement; not null"`
	ReleaseRuleProcessedEventID uint64                    `gorm:"not null"`
	ReleaseRuleProcessedEvent   ReleaseRuleProcessedEvent `gorm:"foreignKey:OrganizationID,ReleaseRuleProcessedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Success                     bool                      `gorm:"not null"`
	CreatedAt                   time.Time                 `gorm:"not null"`
}

type HTTPApiApprovalRuleOutcome struct {
	ApprovalRuleOutcome
	HTTPApiApprovalRuleID uint64              `gorm:"not null"`
	HTTPApiApprovalRule   HTTPApiApprovalRule `gorm:"foreignKey:OrganizationID,HTTPApiApprovalRuleID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ResponseCode          uint8               `gorm:"not null"`
	ResponseContentType   string              `gorm:"not null"`
	ResponseBody          []byte              `gorm:"not null"`
}

type ScheduleApprovalRuleOutcome struct {
	ApprovalRuleOutcome
	ScheduleApprovalRuleID uint64               `gorm:"not null"`
	ScheduleApprovalRule   ScheduleApprovalRule `gorm:"foreignKey:OrganizationID,ScheduleApprovalRuleID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type ManualApprovalRuleOutcome struct {
	ApprovalRuleOutcome
	ManualApprovalRuleID uint64             `gorm:"not null"`
	ManualApprovalRule   ManualApprovalRule `gorm:"foreignKey:OrganizationID,ManualApprovalRuleID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Comments             sql.NullString
}

//
// ******** Find/load functions ********/
//

func FindAllScheduleApprovalRuleOutcomes(db *gorm.DB, organizationID string, releaseID uint64) ([]ScheduleApprovalRuleOutcome, error) {
	var result []ScheduleApprovalRuleOutcome

	tx := db.
		Joins("LEFT JOIN release_rule_processed_events "+
			"ON release_rule_processed_events.organization_id = schedule_approval_rule_outcomes.organization_id "+
			"AND release_rule_processed_events.id = schedule_approval_rule_outcomes.release_rule_processed_event_id").
		Where("schedule_approval_rule_outcomes.organization_id = ? AND release_rule_processed_events.release_id = ?",
			organizationID, releaseID)
	tx = tx.Find(&result)
	return result, tx.Error
}

package dbmodels

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// ApprovalRuleOutcome ...
type ApprovalRuleOutcome struct {
	BaseModel
	ID                                    uint64                              `gorm:"primaryKey; autoIncrement; not null"`
	DeploymentRequestRuleProcessedEventID uint64                              `gorm:"not null"`
	DeploymentRequestRuleProcessedEvent   DeploymentRequestRuleProcessedEvent `gorm:"foreignKey:OrganizationID,DeploymentRequestRuleProcessedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Success                               bool                                `gorm:"not null"`
	CreatedAt                             time.Time                           `gorm:"not null"`
}

// HTTPApiApprovalRuleOutcome ...
type HTTPApiApprovalRuleOutcome struct {
	ApprovalRuleOutcome
	HTTPApiApprovalRuleID uint64              `gorm:"not null"`
	HTTPApiApprovalRule   HTTPApiApprovalRule `gorm:"foreignKey:OrganizationID,HTTPApiApprovalRuleID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ResponseCode          uint8               `gorm:"not null"`
	ResponseContentType   string              `gorm:"not null"`
	ResponseBody          []byte              `gorm:"not null"`
}

// ScheduleApprovalRuleOutcome ...
type ScheduleApprovalRuleOutcome struct {
	ApprovalRuleOutcome
	ScheduleApprovalRuleID uint64               `gorm:"not null"`
	ScheduleApprovalRule   ScheduleApprovalRule `gorm:"foreignKey:OrganizationID,ScheduleApprovalRuleID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ManualApprovalRuleOutcome ...
type ManualApprovalRuleOutcome struct {
	ApprovalRuleOutcome
	ManualApprovalRuleID uint64             `gorm:"not null"`
	ManualApprovalRule   ManualApprovalRule `gorm:"foreignKey:OrganizationID,ManualApprovalRuleID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Comments             sql.NullString
}

// FindAllScheduleApprovalRuleOutcomes ...
func FindAllScheduleApprovalRuleOutcomes(db *gorm.DB, organizationID string, deploymentRequestID uint64) ([]ScheduleApprovalRuleOutcome, error) {
	var result []ScheduleApprovalRuleOutcome

	tx := db.
		Joins("LEFT JOIN deployment_request_rule_processed_events "+
			"ON deployment_request_rule_processed_events.organization_id = schedule_approval_rule_outcomes.organization_id "+
			"AND deployment_request_rule_processed_events.id = schedule_approval_rule_outcomes.deployment_request_rule_processed_event_id").
		Where("schedule_approval_rule_outcomes.organization_id = ? AND deployment_request_rule_processed_events.deployment_request_id = ?",
			organizationID, deploymentRequestID)
	tx = tx.Find(&result)
	return result, tx.Error
}

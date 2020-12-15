package dbmodels

import (
	"database/sql"
	"time"
)

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

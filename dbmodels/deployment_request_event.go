package dbmodels

import (
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
)

// DeploymentRequestEvent ...
type DeploymentRequestEvent struct {
	BaseModel
	ID                  uint64            `gorm:"primaryKey; not null"`
	DeploymentRequestID uint64            `gorm:"not null"`
	ApplicationID       string            `gorm:"type:citext; not null"`
	DeploymentRequest   DeploymentRequest `gorm:"foreignKey:OrganizationID,ApplicationID,DeploymentRequestID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt           time.Time         `gorm:"not null"`
}

// DeploymentRequestCreatedEvent ...
type DeploymentRequestCreatedEvent struct {
	DeploymentRequestEvent
}

// DeploymentRequestCancelledEvent ...
type DeploymentRequestCancelledEvent struct {
	DeploymentRequestEvent
}

// DeploymentRequestRuleProcessedEvent ...
type DeploymentRequestRuleProcessedEvent struct {
	DeploymentRequestEvent
	ResultState  deploymentrequeststate.State `gorm:"type:deployment_request_state; not null"`
	IgnoredError bool                         `gorm:"not null"`
}

package dbmodels

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
)

// ReleaseEvent ...
type ReleaseEvent struct {
	BaseModel
	ID            uint64    `gorm:"primaryKey; not null"`
	ReleaseID     uint64    `gorm:"not null"`
	ApplicationID string    `gorm:"type:citext; not null"`
	Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt     time.Time `gorm:"not null"`
}

// ReleaseCreatedEvent ...
type ReleaseCreatedEvent struct {
	ReleaseEvent
}

// ReleaseCancelledEvent ...
type ReleaseCancelledEvent struct {
	ReleaseEvent
}

// ReleaseRuleProcessedEvent ...
type ReleaseRuleProcessedEvent struct {
	ReleaseEvent
	ResultState  releasestate.State `gorm:"type:release_state; not null"`
	IgnoredError bool               `gorm:"not null"`
}

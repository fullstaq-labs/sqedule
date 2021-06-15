package dbmodels

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
)

//
// ******** Types, constants & variables ********/
//

type ReleaseEvent struct {
	BaseModel
	ID            uint64    `gorm:"primaryKey; not null"`
	ReleaseID     uint64    `gorm:"not null"`
	ApplicationID string    `gorm:"type:citext; not null"`
	Release       Release   `gorm:"foreignKey:OrganizationID,ApplicationID,ReleaseID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	CreatedAt     time.Time `gorm:"not null"`
}

type ReleaseCreatedEvent struct {
	ReleaseEvent
}

type ReleaseCancelledEvent struct {
	ReleaseEvent
}

type ReleaseRuleProcessedEvent struct {
	ReleaseEvent
	ResultState  releasestate.State `gorm:"type:release_state; not null"`
	IgnoredError bool               `gorm:"not null"`
}

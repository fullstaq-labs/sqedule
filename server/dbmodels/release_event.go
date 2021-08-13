package dbmodels

import (
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

type ReleaseEventType string

const (
	ReleaseCreatedEventType       ReleaseEventType = "created"
	ReleaseCancelledEventType     ReleaseEventType = "cancelled"
	ReleaseRuleProcessedEventType ReleaseEventType = "rule_processed"

	NumReleaseEventTypes uint = 3
)

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

	// These are set by LoadReleaseRuleProcessedEventsApprovalRuleOutcomes()
	HTTPApiApprovalRuleOutcome  *HTTPApiApprovalRuleOutcome  `gorm:"-"`
	ScheduleApprovalRuleOutcome *ScheduleApprovalRuleOutcome `gorm:"-"`
	ManualApprovalRuleOutcome   *ManualApprovalRuleOutcome   `gorm:"-"`
}

type ReleaseEventCollection struct {
	ReleaseCreatedEvents       []ReleaseCreatedEvent
	ReleaseCancelledEvents     []ReleaseCancelledEvent
	ReleaseRuleProcessedEvents []ReleaseRuleProcessedEvent
}

//
// ******** ReleaseEventCollection methods ********
//

// NumEvents returns the total number of events in this ReleaseEventCollection.
func (c ReleaseEventCollection) NumEvents() uint {
	return uint(len(c.ReleaseCreatedEvents)) +
		uint(len(c.ReleaseCancelledEvents)) +
		uint(len(c.ReleaseRuleProcessedEvents))
}

//
// ******** Find/load functions ********
//

func FindReleaseEvents(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) (ReleaseEventCollection, error) {
	var result ReleaseEventCollection
	var tx *gorm.DB
	var typesProcessed uint = 0
	var conditions = db.Where("organization_id = ? AND application_id = ? AND release_id = ?", organizationID, applicationID, releaseID)

	typesProcessed++
	tx = db.
		Where(conditions).
		Order("created_at").
		Find(&result.ReleaseCreatedEvents)
	if tx.Error != nil {
		return ReleaseEventCollection{}, tx.Error
	}

	typesProcessed++
	tx = db.
		Where(conditions).
		Order("created_at").
		Find(&result.ReleaseCancelledEvents)
	if tx.Error != nil {
		return ReleaseEventCollection{}, tx.Error
	}

	typesProcessed++
	tx = db.
		Where(conditions).
		Order("created_at").
		Find(&result.ReleaseRuleProcessedEvents)
	if tx.Error != nil {
		return ReleaseEventCollection{}, tx.Error
	}

	if typesProcessed != NumReleaseEventTypes {
		panic("Bug: code does not cover all release event types")
	}

	return result, nil
}

func LoadReleaseRuleProcessedEventsApprovalRuleOutcomes(db *gorm.DB, organizationID string, events []*ReleaseRuleProcessedEvent) error {
	var tx *gorm.DB
	var typesProcessed uint

	eventIDs := CollectReleaseRuleProcessedEventIDs(events)
	eventsIndexByID := indexReleaseRuleProcessedEventsByID(events)
	conditions := db.Where("organization_id = ? AND release_rule_processed_event_id IN ?", organizationID, eventIDs)

	typesProcessed++
	var httpAPIApprovalRuleOutcomes []HTTPApiApprovalRuleOutcome
	tx = db.Where(conditions).Preload("HTTPApiApprovalRule").Find(&httpAPIApprovalRuleOutcomes)
	if tx.Error != nil {
		return tx.Error
	}
	for i := range httpAPIApprovalRuleOutcomes {
		outcome := &httpAPIApprovalRuleOutcomes[i]
		event, ok := eventsIndexByID[outcome.ReleaseRuleProcessedEventID]
		if ok {
			event.HTTPApiApprovalRuleOutcome = outcome
		}
	}

	typesProcessed++
	var scheduleApprovalRuleOutcomes []ScheduleApprovalRuleOutcome
	tx = db.Where(conditions).Preload("ScheduleApprovalRule").Find(&scheduleApprovalRuleOutcomes)
	if tx.Error != nil {
		return tx.Error
	}
	for i := range scheduleApprovalRuleOutcomes {
		outcome := &scheduleApprovalRuleOutcomes[i]
		event, ok := eventsIndexByID[outcome.ReleaseRuleProcessedEventID]
		if ok {
			event.ScheduleApprovalRuleOutcome = outcome
		}
	}

	typesProcessed++
	var manualApprovalRuleOutcomes []ManualApprovalRuleOutcome
	tx = db.Where(conditions).Preload("ManualApprovalRule").Find(&manualApprovalRuleOutcomes)
	if tx.Error != nil {
		return tx.Error
	}
	for i := range manualApprovalRuleOutcomes {
		outcome := &manualApprovalRuleOutcomes[i]
		event, ok := eventsIndexByID[outcome.ReleaseRuleProcessedEventID]
		if ok {
			event.ManualApprovalRuleOutcome = outcome
		}
	}

	if typesProcessed != NumReleaseEventTypes {
		panic("Bug: code does not cover all release event types")
	}

	return nil
}

//
// ******** Other functions ********
//

// MakeReleaseRuleProcessedEventsPointerArray turns a `[]ReleaseRuleProcessedEvent` into a `[]*ReleaseRuleProcessedEvent`.
func MakeReleaseRuleProcessedEventsPointerArray(events []ReleaseRuleProcessedEvent) []*ReleaseRuleProcessedEvent {
	result := make([]*ReleaseRuleProcessedEvent, 0, len(events))
	for i := range events {
		result = append(result, &events[i])
	}
	return result
}

func CollectReleaseRuleProcessedEventIDs(events []*ReleaseRuleProcessedEvent) []uint64 {
	result := make([]uint64, 0, len(events))
	for _, event := range events {
		result = append(result, event.ID)
	}
	return result
}

func indexReleaseRuleProcessedEventsByID(events []*ReleaseRuleProcessedEvent) map[uint64]*ReleaseRuleProcessedEvent {
	result := make(map[uint64]*ReleaseRuleProcessedEvent, len(events))
	for _, event := range events {
		result[event.ID] = event
	}
	return result
}

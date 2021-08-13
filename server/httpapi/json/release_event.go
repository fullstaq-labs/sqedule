package json

import (
	encjson "encoding/json"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ReleaseEventEnum struct {
	*ReleaseCreatedEvent
	*ReleaseCancelledEvent
	*ReleaseRuleProcessedEvent
}

type ReleaseEventBase struct {
	Type      string    `json:"type"`
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type ReleaseCreatedEvent struct {
	ReleaseEventBase
}

type ReleaseCancelledEvent struct {
	ReleaseEventBase
}

type ReleaseRuleProcessedEvent struct {
	ReleaseEventBase
	ResultState         string                  `json:"result_state"`
	IgnoredError        bool                    `json:"ignored_error"`
	ApprovalRuleOutcome ApprovalRuleOutcomeEnum `json:"approval_rule_outcome"`
}

//
// ******** ApprovalRuleOutcomeEnum methods ********
//

// Works around a bug in encoding/json where it doesn't include
// deeply nested anonymous structs.
func (enum ReleaseEventEnum) MarshalJSON() ([]byte, error) {
	if enum.ReleaseCreatedEvent != nil {
		return encjson.Marshal(enum.ReleaseCreatedEvent)
	} else if enum.ReleaseCancelledEvent != nil {
		return encjson.Marshal(enum.ReleaseCancelledEvent)
	} else if enum.ReleaseRuleProcessedEvent != nil {
		return encjson.Marshal(enum.ReleaseRuleProcessedEvent)
	} else {
		panic("Exactly one ReleaseEventEnum field must be set")
	}
}

//
// ******** Constructor functions ********
//

func createReleaseEventBase(theType dbmodels.ReleaseEventType, event dbmodels.ReleaseEvent) ReleaseEventBase {
	return ReleaseEventBase{
		Type:      string(theType),
		ID:        event.ID,
		CreatedAt: event.CreatedAt,
	}
}

func CreateReleaseCreatedEvent(event dbmodels.ReleaseCreatedEvent) ReleaseCreatedEvent {
	return ReleaseCreatedEvent{
		ReleaseEventBase: createReleaseEventBase(dbmodels.ReleaseCreatedEventType, event.ReleaseEvent),
	}
}

func CreateReleaseCancelledEvent(event dbmodels.ReleaseCancelledEvent) ReleaseCancelledEvent {
	return ReleaseCancelledEvent{
		ReleaseEventBase: createReleaseEventBase(dbmodels.ReleaseCancelledEventType, event.ReleaseEvent),
	}
}

func CreateReleaseRuleProcessedEvent(event dbmodels.ReleaseRuleProcessedEvent) ReleaseRuleProcessedEvent {
	return ReleaseRuleProcessedEvent{
		ReleaseEventBase:    createReleaseEventBase(dbmodels.ReleaseRuleProcessedEventType, event.ReleaseEvent),
		ResultState:         string(event.ResultState),
		IgnoredError:        event.IgnoredError,
		ApprovalRuleOutcome: createApprovalRuleOutcomeEnumFromDbmodelsReleaseRuleProcessedEvent(event),
	}
}

func createApprovalRuleOutcomeEnumFromDbmodelsReleaseRuleProcessedEvent(event dbmodels.ReleaseRuleProcessedEvent) ApprovalRuleOutcomeEnum {
	if event.HTTPApiApprovalRuleOutcome != nil {
		outcomeJSON := CreateHTTPApiApprovalRuleOutcome(*event.HTTPApiApprovalRuleOutcome)
		return ApprovalRuleOutcomeEnum{HTTPApiApprovalRuleOutcome: &outcomeJSON}
	}

	if event.ScheduleApprovalRuleOutcome != nil {
		outcomeJSON := CreateScheduleApprovalRuleOutcome(*event.ScheduleApprovalRuleOutcome)
		return ApprovalRuleOutcomeEnum{ScheduleApprovalRuleOutcome: &outcomeJSON}
	}

	if event.ManualApprovalRuleOutcome != nil {
		outcomeJSON := CreateManualApprovalRuleOutcome(*event.ManualApprovalRuleOutcome)
		return ApprovalRuleOutcomeEnum{ManualApprovalRuleOutcome: &outcomeJSON}
	}

	panic("ReleaseRuleProcessedEvent is not associated with an ApprovalRuleOutcome")
}

package json

import (
	"encoding/base64"
	encjson "encoding/json"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApprovalRuleOutcomeEnum struct {
	*HTTPApiApprovalRuleOutcome
	*ScheduleApprovalRuleOutcome
	*ManualApprovalRuleOutcome
}

type ApprovalRuleOutcomeBase struct {
	Type      string    `json:"type"`
	ID        uint64    `json:"id"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

type HTTPApiApprovalRuleOutcome struct {
	ApprovalRuleOutcomeBase
	Rule                HTTPApiApprovalRule `json:"rule"`
	ResponseCode        uint8               `json:"response_code"`
	ResponseContentType string              `json:"response_content_type"`
	ResponseBodyBase64  string              `json:"response_body_base64"`
}

type ScheduleApprovalRuleOutcome struct {
	ApprovalRuleOutcomeBase
	Rule ScheduleApprovalRule `json:"rule"`
}

type ManualApprovalRuleOutcome struct {
	ApprovalRuleOutcomeBase
	Rule     ManualApprovalRule `json:"rule"`
	Comments *string            `json:"comments"`
}

//
// ******** ApprovalRuleOutcomeEnum methods ********
//

// Works around a bug in encoding/json where it doesn't include
// deeply nested anonymous structs.
func (enum ApprovalRuleOutcomeEnum) MarshalJSON() ([]byte, error) {
	if enum.HTTPApiApprovalRuleOutcome != nil {
		return encjson.Marshal(enum.HTTPApiApprovalRuleOutcome)
	} else if enum.ScheduleApprovalRuleOutcome != nil {
		return encjson.Marshal(enum.ScheduleApprovalRuleOutcome)
	} else if enum.ManualApprovalRuleOutcome != nil {
		return encjson.Marshal(enum.ManualApprovalRuleOutcome)
	} else {
		panic("Exactly one ApprovalRuleOutcomeEnum field must be set")
	}
}

//
// ******** Constructor functions ********
//

func createApprovalRuleOutcomeBase(theType dbmodels.ApprovalRuleOutcomeType, outcome dbmodels.ApprovalRuleOutcome) ApprovalRuleOutcomeBase {
	return ApprovalRuleOutcomeBase{
		Type:      string(theType),
		ID:        outcome.ID,
		Success:   outcome.Success,
		CreatedAt: outcome.CreatedAt,
	}
}

func CreateHTTPApiApprovalRuleOutcome(outcome dbmodels.HTTPApiApprovalRuleOutcome) HTTPApiApprovalRuleOutcome {
	return HTTPApiApprovalRuleOutcome{
		ApprovalRuleOutcomeBase: createApprovalRuleOutcomeBase(dbmodels.HTTPApiApprovalRuleOutcomeType, outcome.ApprovalRuleOutcome),
		Rule:                    CreateHTTPApiApprovalRule(outcome.HTTPApiApprovalRule),
		ResponseCode:            outcome.ResponseCode,
		ResponseContentType:     outcome.ResponseContentType,
		ResponseBodyBase64:      base64.StdEncoding.EncodeToString(outcome.ResponseBody),
	}
}

func CreateScheduleApprovalRuleOutcome(outcome dbmodels.ScheduleApprovalRuleOutcome) ScheduleApprovalRuleOutcome {
	return ScheduleApprovalRuleOutcome{
		ApprovalRuleOutcomeBase: createApprovalRuleOutcomeBase(dbmodels.ScheduleApprovalRuleOutcomeType, outcome.ApprovalRuleOutcome),
		Rule:                    CreateScheduleApprovalRule(outcome.ScheduleApprovalRule),
	}
}

func CreateManualApprovalRuleOutcome(outcome dbmodels.ManualApprovalRuleOutcome) ManualApprovalRuleOutcome {
	return ManualApprovalRuleOutcome{
		ApprovalRuleOutcomeBase: createApprovalRuleOutcomeBase(dbmodels.ManualApprovalRuleOutcomeType, outcome.ApprovalRuleOutcome),
		Rule:                    CreateManualApprovalRule(outcome.ManualApprovalRule),
		Comments:                getSqlStringContentsOrNil(outcome.Comments),
	}
}

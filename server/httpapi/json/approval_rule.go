package json

import (
	encjson "encoding/json"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApprovalRuleEnum struct {
	*HTTPApiApprovalRule
	*ScheduleApprovalRule
	*ManualApprovalRule
}

type ApprovalRuleBase struct {
	Type      string    `json:"type"`
	ID        uint64    `json:"id"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

type HTTPApiApprovalRule struct {
	ApprovalRuleBase
	URL              string  `json:"url"`
	Username         *string `json:"username"`
	Password         *string `json:"password"`
	TLSCaCertificate *string `json:"tls_ca_certificate"`
	RetryPolicy      string  `json:"retry_policy"`
	RetryLimit       int     `json:"retry_limit"`
}

type ScheduleApprovalRule struct {
	ApprovalRuleBase
	BeginTime    *string `json:"begin_time"`
	EndTime      *string `json:"end_time"`
	DaysOfWeek   *string `json:"days_of_week"`
	DaysOfMonth  *string `json:"days_of_month"`
	MonthsOfYear *string `json:"months_of_year"`
}

type ManualApprovalRule struct {
	ApprovalRuleBase
	ApprovalPolicy string `json:"approval_policy"`
	Minimum        *int32 `json:"minimum"`
}

//
// ******** ApprovalRuleEnum methods ********
//

// Works around a bug in encoding/json where it doesn't include
// deeply nested anonymous structs.
func (enum ApprovalRuleEnum) MarshalJSON() ([]byte, error) {
	if enum.HTTPApiApprovalRule != nil {
		return encjson.Marshal(enum.HTTPApiApprovalRule)
	} else if enum.ScheduleApprovalRule != nil {
		return encjson.Marshal(enum.ScheduleApprovalRule)
	} else if enum.ManualApprovalRule != nil {
		return encjson.Marshal(enum.ManualApprovalRule)
	} else {
		panic("Exactly one ApprovalRuleEnum field must be set")
	}
}

//
// ******** Constructor functions ********
//

func createApprovalRuleBase(theType dbmodels.ApprovalRuleType, rule dbmodels.ApprovalRule) ApprovalRuleBase {
	return ApprovalRuleBase{
		Type:      string(theType),
		ID:        rule.ID,
		Enabled:   lib.DerefBoolPtrWithDefault(rule.Enabled, true),
		CreatedAt: rule.CreatedAt,
	}
}

func CreateHTTPApiApprovalRule(rule dbmodels.HTTPApiApprovalRule) HTTPApiApprovalRule {
	return HTTPApiApprovalRule{
		ApprovalRuleBase: createApprovalRuleBase(dbmodels.HTTPApiApprovalRuleType, rule.ApprovalRule),
		URL:              rule.URL,
		Username:         getSqlStringContentsOrNil(rule.Username),
		TLSCaCertificate: getSqlStringContentsOrNil(rule.TLSCaCertificate),
		RetryPolicy:      string(rule.RetryPolicy),
		RetryLimit:       rule.RetryLimit,
	}
}

func CreateScheduleApprovalRule(rule dbmodels.ScheduleApprovalRule) ScheduleApprovalRule {
	return ScheduleApprovalRule{
		ApprovalRuleBase: createApprovalRuleBase(dbmodels.ScheduleApprovalRuleType, rule.ApprovalRule),
		BeginTime:        getSqlStringContentsOrNil(rule.BeginTime),
		EndTime:          getSqlStringContentsOrNil(rule.EndTime),
		DaysOfWeek:       getSqlStringContentsOrNil(rule.DaysOfWeek),
		DaysOfMonth:      getSqlStringContentsOrNil(rule.DaysOfMonth),
		MonthsOfYear:     getSqlStringContentsOrNil(rule.MonthsOfYear),
	}
}

func CreateManualApprovalRule(rule dbmodels.ManualApprovalRule) ManualApprovalRule {
	result := ManualApprovalRule{
		ApprovalRuleBase: createApprovalRuleBase(dbmodels.ManualApprovalRuleType, rule.ApprovalRule),
		ApprovalPolicy:   string(rule.ApprovalPolicy),
	}
	if rule.Minimum.Valid {
		result.Minimum = &rule.Minimum.Int32
	}
	return result
}

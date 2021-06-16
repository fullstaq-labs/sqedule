package json

import (
	"encoding/json"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApprovalRule struct {
	Type dbmodels.ApprovalRuleType
	dbmodels.HTTPApiApprovalRule
	dbmodels.ScheduleApprovalRule
	dbmodels.ManualApprovalRule
}

//
// ******** ApprovalRule methods ********
//

func (rule ApprovalRule) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type": string(rule.Type),
	}

	switch rule.Type {
	case dbmodels.HTTPApiApprovalRuleType:
		generateJSONFieldsForHTTPApiApprovalRule(rule.HTTPApiApprovalRule, result)
	case dbmodels.ScheduleApprovalRuleType:
		generateJSONFieldsForScheduleApprovalRule(rule.ScheduleApprovalRule, result)
	case dbmodels.ManualApprovalRuleType:
		generateJSONFieldsForManualApprovalRule(rule.ManualApprovalRule, result)
	default:
		panic("Unsupported approval rule type " + rule.Type)
	}

	return json.Marshal(result)
}

func generateJSONFieldsForApprovalRuleBase(model dbmodels.ApprovalRule, result map[string]interface{}) {
	result["id"] = model.ID
	result["enabled"] = model.Enabled
	result["created_at"] = model.CreatedAt
}

func generateJSONFieldsForHTTPApiApprovalRule(model dbmodels.HTTPApiApprovalRule, result map[string]interface{}) {
	generateJSONFieldsForApprovalRuleBase(model.ApprovalRule, result)
	result["url"] = model.URL
	result["username"] = getSqlStringContentsOrNil(model.Username)
	result["tls_ca_certificate"] = getSqlStringContentsOrNil(model.TLSCaCertificate)
	result["retry_policy"] = string(model.RetryPolicy)
	result["retry_limit"] = model.RetryLimit
}

func generateJSONFieldsForScheduleApprovalRule(model dbmodels.ScheduleApprovalRule, result map[string]interface{}) {
	generateJSONFieldsForApprovalRuleBase(model.ApprovalRule, result)
	result["begin_time"] = getSqlStringContentsOrNil(model.BeginTime)
	result["end_time"] = getSqlStringContentsOrNil(model.EndTime)
	result["days_of_week"] = getSqlStringContentsOrNil(model.DaysOfWeek)
	result["days_of_month"] = getSqlStringContentsOrNil(model.DaysOfMonth)
	result["months_of_year"] = getSqlStringContentsOrNil(model.MonthsOfYear)
}

func generateJSONFieldsForManualApprovalRule(model dbmodels.ManualApprovalRule, result map[string]interface{}) {
	generateJSONFieldsForApprovalRuleBase(model.ApprovalRule, result)
	result["approval_policy"] = string(model.ApprovalPolicy)
	if model.Minimum.Valid {
		result["minimum"] = model.Minimum.Int32
	} else {
		result["minimum"] = nil
	}
}

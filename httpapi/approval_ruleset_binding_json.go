package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type approvalRulesetBindingJSON struct {
	Application     *applicationJSON    `json:"application,omitempty"`
	ApprovalRuleset approvalRulesetJSON `json:"approval_ruleset"`
	Mode            string              `json:"mode"`
}

func createApprovalRulesetBindingJSONFromDbModel(binding dbmodels.ApprovalRulesetBinding, includeApplication bool) approvalRulesetBindingJSON {
	result := approvalRulesetBindingJSON{
		ApprovalRuleset: createApprovalRulesetJSONFromDbModel(binding.ApprovalRuleset),
		Mode:            string(binding.Mode),
	}
	if includeApplication {
		appJSON := createApplicationJSONFromDbModel(binding.Application)
		result.Application = &appJSON
	}
	return result
}

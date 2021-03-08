package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type releaseApprovalRulesetBindingJSON struct {
	ApprovalRuleset approvalRulesetJSON `json:"approval_ruleset"`
	Mode            string              `json:"mode"`
}

func createReleaseApprovalRulesetBindingJSONFromDbModel(binding dbmodels.ReleaseApprovalRulesetBinding) releaseApprovalRulesetBindingJSON {
	result := releaseApprovalRulesetBindingJSON{
		ApprovalRuleset: createApprovalRulesetJSONFromDbModel(binding.ApprovalRuleset,
			binding.ApprovalRulesetMajorVersion, binding.ApprovalRulesetMinorVersion),
		Mode: string(binding.Mode),
	}
	return result
}

package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type releaseApprovalRulesetBindingJSON struct {
	Mode string `json:"mode"`
}

type releaseApprovalRulesetBindingWithRulesetAssociationJSON struct {
	releaseApprovalRulesetBindingJSON
	ApprovalRuleset approvalRulesetJSON `json:"approval_ruleset"`
}

type releaseApprovalRulesetBindingWithReleaseAssociationJSON struct {
	releaseApprovalRulesetBindingJSON
	Release releaseWithApplicationAssociationJSON `json:"release"`
}

func createReleaseApprovalRulesetBindingJSONFromDbModel(binding dbmodels.ReleaseApprovalRulesetBinding) releaseApprovalRulesetBindingJSON {
	return releaseApprovalRulesetBindingJSON{
		Mode: string(binding.Mode),
	}
}

func createReleaseApprovalRulesetBindingWithReleaseAssociationJSONFromDbModel(binding dbmodels.ReleaseApprovalRulesetBinding) releaseApprovalRulesetBindingWithReleaseAssociationJSON {
	return releaseApprovalRulesetBindingWithReleaseAssociationJSON{
		releaseApprovalRulesetBindingJSON: createReleaseApprovalRulesetBindingJSONFromDbModel(binding),
		Release:                           createReleaseWithApplicationAssociationJSONFromDbModel(binding.Release),
	}
}

func createReleaseApprovalRulesetBindingWithRulesetAssociationJSONFromDbModel(binding dbmodels.ReleaseApprovalRulesetBinding) releaseApprovalRulesetBindingWithRulesetAssociationJSON {
	return releaseApprovalRulesetBindingWithRulesetAssociationJSON{
		releaseApprovalRulesetBindingJSON: createReleaseApprovalRulesetBindingJSONFromDbModel(binding),
		ApprovalRuleset: createApprovalRulesetJSONFromDbModel(binding.ApprovalRuleset,
			binding.ApprovalRulesetMajorVersion, binding.ApprovalRulesetMinorVersion),
	}
}

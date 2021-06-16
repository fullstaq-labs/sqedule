package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ReleaseApprovalRulesetBinding struct {
	Mode string `json:"mode"`
}

type ReleaseApprovalRulesetBindingWithRulesetAssociation struct {
	ReleaseApprovalRulesetBinding
	ApprovalRuleset ApprovalRulesetWithLatestApprovedVersion `json:"approval_ruleset"`
}

type ReleaseApprovalRulesetBindingWithReleaseAssociation struct {
	ReleaseApprovalRulesetBinding
	Release ReleaseWithApplicationAssociation `json:"release"`
}

//
// ******** Constructor functions ********
//

func CreateFromDbReleaseApprovalRulesetBinding(binding dbmodels.ReleaseApprovalRulesetBinding) ReleaseApprovalRulesetBinding {
	return ReleaseApprovalRulesetBinding{
		Mode: string(binding.Mode),
	}
}

func CreateFromDbReleaseApprovalRulesetBindingWithReleaseAssociation(binding dbmodels.ReleaseApprovalRulesetBinding) ReleaseApprovalRulesetBindingWithReleaseAssociation {
	return ReleaseApprovalRulesetBindingWithReleaseAssociation{
		ReleaseApprovalRulesetBinding: CreateFromDbReleaseApprovalRulesetBinding(binding),
		Release:                       CreateFromDbReleaseWithApplicationAssociation(binding.Release),
	}
}

func CreateFromDbReleaseApprovalRulesetBindingWithRulesetAssociation(binding dbmodels.ReleaseApprovalRulesetBinding) ReleaseApprovalRulesetBindingWithRulesetAssociation {
	return ReleaseApprovalRulesetBindingWithRulesetAssociation{
		ReleaseApprovalRulesetBinding: CreateFromDbReleaseApprovalRulesetBinding(binding),
		ApprovalRuleset: CreateApprovalRulesetWithLatestApprovedVersion(binding.ApprovalRuleset,
			binding.ApprovalRulesetVersion, binding.ApprovalRulesetAdjustment),
	}
}

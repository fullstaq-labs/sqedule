package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type ReleaseApprovalRulesetBinding struct {
	Mode string `json:"mode"`
}

type ReleaseApprovalRulesetBindingWithRulesetAssociation struct {
	ReleaseApprovalRulesetBinding
	ApprovalRuleset ApprovalRuleset `json:"approval_ruleset"`
}

type ReleaseApprovalRulesetBindingWithReleaseAssociation struct {
	ReleaseApprovalRulesetBinding
	Release ReleaseWithApplicationAssociation `json:"release"`
}

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
		ApprovalRuleset: CreateFromDbApprovalRuleset(binding.ApprovalRuleset,
			binding.ApprovalRulesetMajorVersion, binding.ApprovalRulesetMinorVersion),
	}
}

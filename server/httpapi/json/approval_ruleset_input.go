package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type ApprovalRulesetInput struct {
	ID      *string                      `json:"id"`
	Version *ApprovalRulesetVersionInput `json:"version"`
}

type ApprovalRulesetVersionInput struct {
	ReviewableVersionInputBase
	DisplayName   *string              `json:"display_name"`
	Description   *string              `json:"description"`
	Enabled       *bool                `json:"enabled"`
	ApprovalRules *[]ApprovalRuleInput `json:"approval_rules"`
}

func (input ApprovalRulesetVersionInput) ToDbmodelsApprovalRulesetContents(organizationID string) dbmodels.ApprovalRulesetContents {
	var contents dbmodels.ApprovalRulesetContents
	if input.ApprovalRules == nil {
		return contents
	}

	for _, input := range *input.ApprovalRules {
		input.AppendToDbmodelsApprovalRulesetContents(organizationID, &contents)
	}
	return contents
}

func PatchApprovalRuleset(ruleset *dbmodels.ApprovalRuleset, input ApprovalRulesetInput) {
	if input.ID != nil {
		ruleset.ID = *input.ID
	}
}

func PatchApprovalRulesetAdjustment(organizationID string, adjustment *dbmodels.ApprovalRulesetAdjustment, rules *dbmodels.ApprovalRulesetContents, input ApprovalRulesetVersionInput) {
	if input.DisplayName != nil {
		adjustment.DisplayName = *input.DisplayName
	}
	if input.Description != nil {
		adjustment.Description = *input.Description
	}
	if input.Enabled != nil {
		adjustment.Enabled = *input.Enabled
	}
	if input.ApprovalRules != nil {
		*rules = input.ToDbmodelsApprovalRulesetContents(organizationID)
	}
	rules.ForEach(func(rule dbmodels.IApprovalRule) error {
		rule.AssociateWithApprovalRulesetAdjustment(*adjustment)
		return nil
	})
}

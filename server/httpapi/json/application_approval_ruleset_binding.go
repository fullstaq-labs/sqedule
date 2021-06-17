package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApplicationApprovalRulesetBinding struct {
	VersionNumber    *uint32 `json:"version_number"`
	AdjustmentNumber uint32  `json:"adjustment_number"`
	Mode             string  `json:"mode"`
}

type ApplicationApprovalRulesetBindingWithApplicationAssociation struct {
	ApplicationApprovalRulesetBinding
	Application Application `json:"application"`
}

type ApplicationApprovalRulesetBindingWithRulesetAssociation struct {
	ApplicationApprovalRulesetBinding
	ApprovalRuleset ApprovalRulesetWithLatestApprovedVersion `json:"approval_ruleset"`
}

//
// ******** Constructor functions ********
//

func CreateFromDbApplicationApprovalRulesetBinding(binding dbmodels.ApplicationApprovalRulesetBinding, version dbmodels.ApplicationApprovalRulesetBindingVersion,
	adjustment dbmodels.ApplicationApprovalRulesetBindingAdjustment) ApplicationApprovalRulesetBinding {

	return ApplicationApprovalRulesetBinding{
		VersionNumber:    version.VersionNumber,
		AdjustmentNumber: adjustment.AdjustmentNumber,
		Mode:             string(adjustment.Mode),
	}
}

func CreateFromDbApplicationApprovalRulesetBindingWithApplicationAssociation(binding dbmodels.ApplicationApprovalRulesetBinding, version dbmodels.ApplicationApprovalRulesetBindingVersion,
	adjustment dbmodels.ApplicationApprovalRulesetBindingAdjustment) ApplicationApprovalRulesetBindingWithApplicationAssociation {

	if binding.Application.Version == nil {
		panic("Associated application must have an associated version")
	}
	if binding.Application.Version.VersionNumber == nil {
		panic("Associated application's associated version must be finalized")
	}
	if binding.Application.Version.Adjustment == nil {
		panic("Associated application must have an associated adjustment")
	}

	return ApplicationApprovalRulesetBindingWithApplicationAssociation{
		ApplicationApprovalRulesetBinding: CreateFromDbApplicationApprovalRulesetBinding(binding, version, adjustment),
		Application: CreateFromDbApplication(binding.Application, *binding.Application.Version,
			*binding.Application.Version.Adjustment, nil),
	}
}

func CreateFromDbApplicationApprovalRulesetBindingWithRulesetAssociation(binding dbmodels.ApplicationApprovalRulesetBinding, version dbmodels.ApplicationApprovalRulesetBindingVersion,
	adjustment dbmodels.ApplicationApprovalRulesetBindingAdjustment) ApplicationApprovalRulesetBindingWithRulesetAssociation {

	if binding.ApprovalRuleset.Version == nil {
		panic("Given approval ruleset must have an associated version")
	}
	if binding.ApprovalRuleset.Version.VersionNumber == nil {
		panic("Given approval ruleset's associated version must be finalized")
	}
	if binding.ApprovalRuleset.Version.Adjustment == nil {
		panic("Given approval ruleset must have an associated adjustment")
	}

	return ApplicationApprovalRulesetBindingWithRulesetAssociation{
		ApplicationApprovalRulesetBinding: CreateFromDbApplicationApprovalRulesetBinding(binding, version, adjustment),
		ApprovalRuleset: CreateApprovalRulesetWithLatestApprovedVersion(binding.ApprovalRuleset,
			*binding.ApprovalRuleset.Version, *binding.ApprovalRuleset.Version.Adjustment),
	}
}

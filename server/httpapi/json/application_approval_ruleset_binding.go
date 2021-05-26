package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

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

	if binding.Application.LatestVersion == nil {
		panic("Associated application must have an associated latest version")
	}
	if binding.Application.LatestVersion.VersionNumber == nil {
		panic("Associated application's associated latest version must be finalized")
	}
	if binding.Application.LatestAdjustment == nil {
		panic("Associated application must have an associated latest adjustment")
	}

	return ApplicationApprovalRulesetBindingWithApplicationAssociation{
		ApplicationApprovalRulesetBinding: CreateFromDbApplicationApprovalRulesetBinding(binding, version, adjustment),
		Application: CreateFromDbApplication(binding.Application, *binding.Application.LatestVersion,
			*binding.Application.LatestAdjustment, nil),
	}
}

func CreateFromDbApplicationApprovalRulesetBindingWithRulesetAssociation(binding dbmodels.ApplicationApprovalRulesetBinding, version dbmodels.ApplicationApprovalRulesetBindingVersion,
	adjustment dbmodels.ApplicationApprovalRulesetBindingAdjustment) ApplicationApprovalRulesetBindingWithRulesetAssociation {

	if binding.ApprovalRuleset.LatestVersion == nil {
		panic("Given approval ruleset must have an associated latest version")
	}
	if binding.ApprovalRuleset.LatestVersion.VersionNumber == nil {
		panic("Given approval ruleset's associated latest version must be finalized")
	}
	if binding.ApprovalRuleset.LatestAdjustment == nil {
		panic("Given approval ruleset must have an associated latest adjustment")
	}

	return ApplicationApprovalRulesetBindingWithRulesetAssociation{
		ApplicationApprovalRulesetBinding: CreateFromDbApplicationApprovalRulesetBinding(binding, version, adjustment),
		ApprovalRuleset: CreateApprovalRulesetWithLatestApprovedVersion(binding.ApprovalRuleset,
			*binding.ApprovalRuleset.LatestVersion, *binding.ApprovalRuleset.LatestAdjustment),
	}
}

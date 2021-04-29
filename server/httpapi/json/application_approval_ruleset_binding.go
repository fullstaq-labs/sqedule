package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type ApplicationApprovalRulesetBinding struct {
	MajorVersionNumber *uint32 `json:"major_version_number"`
	MinorVersionNumber uint32  `json:"minor_version_number"`
	Mode               string  `json:"mode"`
}

type ApplicationApprovalRulesetBindingWithApplicationAssociation struct {
	ApplicationApprovalRulesetBinding
	Application Application `json:"application"`
}

type ApplicationApprovalRulesetBindingWithRulesetAssociation struct {
	ApplicationApprovalRulesetBinding
	ApprovalRuleset ApprovalRuleset `json:"approval_ruleset"`
}

func CreateFromDbApplicationApprovalRulesetBinding(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion) ApplicationApprovalRulesetBinding {

	return ApplicationApprovalRulesetBinding{
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		Mode:               string(minorVersion.Mode),
	}
}

func CreateFromDbApplicationApprovalRulesetBindingWithApplicationAssociation(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion) ApplicationApprovalRulesetBindingWithApplicationAssociation {

	if binding.Application.LatestMajorVersion == nil {
		panic("Associated application must have an associated latest major version")
	}
	if binding.Application.LatestMajorVersion.VersionNumber == nil {
		panic("Associated application's associated latest major version must be finalized")
	}
	if binding.Application.LatestMinorVersion == nil {
		panic("Associated application must have an associated latest minor version")
	}

	return ApplicationApprovalRulesetBindingWithApplicationAssociation{
		ApplicationApprovalRulesetBinding: CreateFromDbApplicationApprovalRulesetBinding(binding, majorVersion, minorVersion),
		Application: CreateFromDbApplication(binding.Application, *binding.Application.LatestMajorVersion,
			*binding.Application.LatestMinorVersion, nil),
	}
}

func CreateFromDbApplicationApprovalRulesetBindingWithRulesetAssociation(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion) ApplicationApprovalRulesetBindingWithRulesetAssociation {

	if binding.ApprovalRuleset.LatestMajorVersion == nil {
		panic("Given approval ruleset must have an associated latest major version")
	}
	if binding.ApprovalRuleset.LatestMajorVersion.VersionNumber == nil {
		panic("Given approval ruleset's associated latest major version must be finalized")
	}
	if binding.ApprovalRuleset.LatestMinorVersion == nil {
		panic("Given approval ruleset must have an associated latest minor version")
	}

	return ApplicationApprovalRulesetBindingWithRulesetAssociation{
		ApplicationApprovalRulesetBinding: CreateFromDbApplicationApprovalRulesetBinding(binding, majorVersion, minorVersion),
		ApprovalRuleset: CreateFromDbApprovalRuleset(binding.ApprovalRuleset,
			*binding.ApprovalRuleset.LatestMajorVersion, *binding.ApprovalRuleset.LatestMinorVersion),
	}
}

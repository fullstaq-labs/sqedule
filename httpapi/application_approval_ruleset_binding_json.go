package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type applicationApprovalRulesetBindingJSON struct {
	MajorVersionNumber *uint32 `json:"major_version_number"`
	MinorVersionNumber uint32  `json:"minor_version_number"`
	Mode               string  `json:"mode"`
}

type applicationApprovalRulesetBindingWithApplicationAssociationJSON struct {
	applicationApprovalRulesetBindingJSON
	Application applicationJSON `json:"application"`
}

type applicationApprovalRulesetBindingWithRulesetAssociationJSON struct {
	applicationApprovalRulesetBindingJSON
	ApprovalRuleset approvalRulesetJSON `json:"approval_ruleset"`
}

func createApplicationApprovalRulesetBindingJSONFromDbModel(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion) applicationApprovalRulesetBindingJSON {

	return applicationApprovalRulesetBindingJSON{
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		Mode:               string(minorVersion.Mode),
	}
}

func createApplicationApprovalRulesetBindingWithApplicationAssociationJSONFromDbModel(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion) applicationApprovalRulesetBindingWithApplicationAssociationJSON {

	if binding.Application.LatestMajorVersion == nil {
		panic("Associated application must have an associated latest major version")
	}
	if binding.Application.LatestMajorVersion.VersionNumber == nil {
		panic("Associated application's associated latest major version must be finalized")
	}
	if binding.Application.LatestMinorVersion == nil {
		panic("Associated application must have an associated latest minor version")
	}

	return applicationApprovalRulesetBindingWithApplicationAssociationJSON{
		applicationApprovalRulesetBindingJSON: createApplicationApprovalRulesetBindingJSONFromDbModel(binding, majorVersion, minorVersion),
		Application: createApplicationJSONFromDbModel(binding.Application, *binding.Application.LatestMajorVersion,
			*binding.Application.LatestMinorVersion, nil),
	}
}

func createApplicationApprovalRulesetBindingWithRulesetAssociationJSONFromDbModel(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion) applicationApprovalRulesetBindingWithRulesetAssociationJSON {

	if binding.ApprovalRuleset.LatestMajorVersion == nil {
		panic("Given approval ruleset must have an associated latest major version")
	}
	if binding.ApprovalRuleset.LatestMajorVersion.VersionNumber == nil {
		panic("Given approval ruleset's associated latest major version must be finalized")
	}
	if binding.ApprovalRuleset.LatestMinorVersion == nil {
		panic("Given approval ruleset must have an associated latest minor version")
	}

	return applicationApprovalRulesetBindingWithRulesetAssociationJSON{
		applicationApprovalRulesetBindingJSON: createApplicationApprovalRulesetBindingJSONFromDbModel(binding, majorVersion, minorVersion),
		ApprovalRuleset: createApprovalRulesetJSONFromDbModel(binding.ApprovalRuleset,
			*binding.ApprovalRuleset.LatestMajorVersion, *binding.ApprovalRuleset.LatestMinorVersion),
	}
}

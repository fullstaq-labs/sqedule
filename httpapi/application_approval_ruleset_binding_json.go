package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type applicationApprovalRulesetBindingJSON struct {
	MajorVersionNumber *uint32             `json:"major_version_number"`
	MinorVersionNumber uint32              `json:"minor_version_number"`
	Application        *applicationJSON    `json:"application,omitempty"`
	ApprovalRuleset    approvalRulesetJSON `json:"approval_ruleset"`
	Mode               string              `json:"mode"`
}

func createApplicationApprovalRulesetBindingJSONFromDbModel(binding dbmodels.ApplicationApprovalRulesetBinding, majorVersion dbmodels.ApplicationApprovalRulesetBindingMajorVersion,
	minorVersion dbmodels.ApplicationApprovalRulesetBindingMinorVersion, includeApplication bool) applicationApprovalRulesetBindingJSON {

	if binding.ApprovalRuleset.LatestMajorVersion == nil {
		panic("Given approval ruleset must have an associated latest major version")
	}
	if binding.ApprovalRuleset.LatestMajorVersion.VersionNumber == nil {
		panic("Given approval ruleset's associated latest major version must be finalized")
	}
	if binding.ApprovalRuleset.LatestMinorVersion == nil {
		panic("Given approval ruleset must have an associated latest minor version")
	}

	result := applicationApprovalRulesetBindingJSON{
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		ApprovalRuleset: createApprovalRulesetJSONFromDbModel(binding.ApprovalRuleset,
			*binding.ApprovalRuleset.LatestMajorVersion, *binding.ApprovalRuleset.LatestMinorVersion),
		Mode: string(minorVersion.Mode),
	}
	if includeApplication {
		if binding.Application.LatestMajorVersion == nil {
			panic("Associated application must have an associated latest major version")
		}
		if binding.Application.LatestMajorVersion.VersionNumber == nil {
			panic("Associated application's associated latest major version must be finalized")
		}
		if binding.Application.LatestMinorVersion == nil {
			panic("Associated application must have an associated latest minor version")
		}
		appJSON := createApplicationJSONFromDbModel(binding.Application, *binding.Application.LatestMajorVersion,
			*binding.Application.LatestMinorVersion, nil)
		result.Application = &appJSON
	}
	return result
}

package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type applicationApprovalRulesetBindingJSON struct {
	LatestMajorVersionNumber uint32              `json:"latest_major_version_number"`
	LatestMinorVersionNumber uint32              `json:"latest_minor_version_number"`
	Application              *applicationJSON    `json:"application,omitempty"`
	ApprovalRuleset          approvalRulesetJSON `json:"approval_ruleset"`
	Mode                     string              `json:"mode"`
}

func createApplicationApprovalRulesetBindingJSONFromDbModel(binding dbmodels.ApplicationApprovalRulesetBinding, includeApplication bool) applicationApprovalRulesetBindingJSON {
	if binding.LatestMajorVersion == nil {
		panic("Given binding must have an associated latest major version")
	}
	if binding.LatestMajorVersion.VersionNumber == nil {
		panic("Given binding's associated latest major version must be finalized")
	}
	if binding.LatestMinorVersion == nil {
		panic("Given binding must have an associated latest minor version")
	}

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
		LatestMajorVersionNumber: *binding.LatestMajorVersion.VersionNumber,
		LatestMinorVersionNumber: binding.LatestMinorVersion.VersionNumber,
		ApprovalRuleset: createApprovalRulesetJSONFromDbModel(binding.ApprovalRuleset,
			*binding.ApprovalRuleset.LatestMajorVersion, *binding.ApprovalRuleset.LatestMinorVersion),
		Mode: string(binding.LatestMinorVersion.Mode),
	}
	if includeApplication {
		appJSON := createApplicationJSONFromDbModel(binding.Application)
		result.Application = &appJSON
	}
	return result
}

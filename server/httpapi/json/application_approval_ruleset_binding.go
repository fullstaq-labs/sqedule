package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApplicationApprovalRulesetBindingBase struct {
	Application     *ApplicationWithLatestApprovedVersion     `json:"application,omitempty"`
	ApprovalRuleset *ApprovalRulesetWithLatestApprovedVersion `json:"approval_ruleset,omitempty"`
}

type ApplicationApprovalRulesetBindingWithVersion struct {
	ReviewableBase
	ApplicationApprovalRulesetBindingBase
	Version *ApplicationApprovalRulesetBindingVersion `json:"version"`
}

type ApplicationApprovalRulesetBindingWithLatestApprovedVersion struct {
	ReviewableBase
	ApplicationApprovalRulesetBindingBase
	LatestApprovedVersion *ApplicationApprovalRulesetBindingVersion `json:"latest_approved_version"`
}

type ApplicationApprovalRulesetBindingVersion struct {
	ReviewableVersionBase
	Mode string `json:"mode"`
}

//
// ******** Constructor functions ********
//

func CreateApplicationApprovalRulesetBindingVersion(version dbmodels.ApplicationApprovalRulesetBindingVersion) ApplicationApprovalRulesetBindingVersion {
	return ApplicationApprovalRulesetBindingVersion{
		ReviewableVersionBase: createReviewableVersionBase(version.ReviewableVersionBase, version.Adjustment.ReviewableAdjustmentBase),
		Mode:                  string(version.Adjustment.Mode),
	}
}

func CreateApplicationApprovalRulesetBindingWithLatestApprovedVersion(binding dbmodels.ApplicationApprovalRulesetBinding,
	version *dbmodels.ApplicationApprovalRulesetBindingVersion) ApplicationApprovalRulesetBindingWithLatestApprovedVersion {

	var versionJSON *ApplicationApprovalRulesetBindingVersion

	if version != nil {
		versionJSONStruct := CreateApplicationApprovalRulesetBindingVersion(*version)
		versionJSON = &versionJSONStruct
	}

	return ApplicationApprovalRulesetBindingWithLatestApprovedVersion{
		ReviewableBase:        createReviewableBase(binding.ReviewableBase),
		LatestApprovedVersion: versionJSON,
	}
}

func CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(binding dbmodels.ApplicationApprovalRulesetBinding,
	version *dbmodels.ApplicationApprovalRulesetBindingVersion, includeApp bool, includeRuleset bool) ApplicationApprovalRulesetBindingWithLatestApprovedVersion {

	result := CreateApplicationApprovalRulesetBindingWithLatestApprovedVersion(binding, version)
	if includeApp {
		appJSONStruct := CreateApplicationWithLatestApprovedVersion(binding.Application, binding.Application.Version)
		result.ApplicationApprovalRulesetBindingBase.Application = &appJSONStruct
	}
	if includeRuleset {
		rulesetJSONStruct := CreateApprovalRulesetWithLatestApprovedVersion(binding.ApprovalRuleset, binding.ApprovalRuleset.Version)
		result.ApplicationApprovalRulesetBindingBase.ApprovalRuleset = &rulesetJSONStruct
	}
	return result
}

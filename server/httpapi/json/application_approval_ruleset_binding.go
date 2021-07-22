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
// ******** Constructor functions: Version ********
//

func CreateApplicationApprovalRulesetBindingVersion(version dbmodels.ApplicationApprovalRulesetBindingVersion) ApplicationApprovalRulesetBindingVersion {
	return ApplicationApprovalRulesetBindingVersion{
		ReviewableVersionBase: createReviewableVersionBase(version.ReviewableVersionBase, version.Adjustment.ReviewableAdjustmentBase),
		Mode:                  string(version.Adjustment.Mode),
	}
}

//
// ******** Constructor functions: WithVersion ********
//

func CreateApplicationApprovalRulesetBindingWithVersion(binding dbmodels.ApplicationApprovalRulesetBinding, version *dbmodels.ApplicationApprovalRulesetBindingVersion) ApplicationApprovalRulesetBindingWithVersion {
	result := ApplicationApprovalRulesetBindingWithVersion{
		ReviewableBase: createReviewableBase(binding.ReviewableBase),
	}
	if version != nil {
		jsonStruct := CreateApplicationApprovalRulesetBindingVersion(*version)
		result.Version = &jsonStruct
	}
	return result
}

func CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding dbmodels.ApplicationApprovalRulesetBinding, version *dbmodels.ApplicationApprovalRulesetBindingVersion,
	includeApp bool, includeRuleset bool) ApplicationApprovalRulesetBindingWithVersion {

	result := CreateApplicationApprovalRulesetBindingWithVersion(binding, version)
	if includeApp {
		jsonStruct := CreateApplicationWithLatestApprovedVersion(binding.Application, binding.Application.Version)
		result.ApplicationApprovalRulesetBindingBase.Application = &jsonStruct
	}
	if includeRuleset {
		jsonStruct := CreateApprovalRulesetWithLatestApprovedVersion(binding.ApprovalRuleset, binding.ApprovalRuleset.Version)
		result.ApplicationApprovalRulesetBindingBase.ApprovalRuleset = &jsonStruct
	}
	return result
}

//
// ******** Constructor functions: WithLatestApprovedVersion ********
//

func CreateApplicationApprovalRulesetBindingWithLatestApprovedVersion(binding dbmodels.ApplicationApprovalRulesetBinding,
	version *dbmodels.ApplicationApprovalRulesetBindingVersion) ApplicationApprovalRulesetBindingWithLatestApprovedVersion {

	result := ApplicationApprovalRulesetBindingWithLatestApprovedVersion{
		ReviewableBase: createReviewableBase(binding.ReviewableBase),
	}
	if version != nil {
		jsonStruct := CreateApplicationApprovalRulesetBindingVersion(*version)
		result.LatestApprovedVersion = &jsonStruct
	}
	return result
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

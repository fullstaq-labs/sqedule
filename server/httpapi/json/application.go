package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApplicationBase struct {
	ID                      string                                                        `json:"id"`
	ApprovalRulesetBindings *[]ApplicationApprovalRulesetBindingWithLatestApprovedVersion `json:"approval_ruleset_bindings,omitempty"`
}

type ApplicationWithVersion struct {
	ReviewableBase
	ApplicationBase
	Version *ApplicationVersion `json:"version"`
}

type ApplicationWithLatestApprovedVersion struct {
	ReviewableBase
	ApplicationBase
	LatestApprovedVersion *ApplicationVersion `json:"latest_approved_version"`
}

type ApplicationVersion struct {
	ReviewableVersionBase
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

//
// ******** Constructor functions: Version ********
//

func CreateApplicationVersion(version dbmodels.ApplicationVersion) ApplicationVersion {
	return ApplicationVersion{
		ReviewableVersionBase: createReviewableVersionBase(version.ReviewableVersionBase, version.Adjustment.ReviewableAdjustmentBase),
		DisplayName:           version.Adjustment.DisplayName,
		Enabled:               version.Adjustment.IsEnabled(),
	}
}

//
// ******** Constructor functions: WithVersion ********
//

func CreateApplicationWithVersion(app dbmodels.Application, version *dbmodels.ApplicationVersion) ApplicationWithVersion {
	result := ApplicationWithVersion{
		ReviewableBase: createReviewableBase(app.ReviewableBase),
		ApplicationBase: ApplicationBase{
			ID: app.ID,
		},
	}
	if version != nil {
		jsonStruct := CreateApplicationVersion(*version)
		result.Version = &jsonStruct
	}
	return result
}

func CreateApplicationWithVersionAndAssociations(app dbmodels.Application, version *dbmodels.ApplicationVersion,
	rulesetBindings *[]dbmodels.ApplicationApprovalRulesetBinding) ApplicationWithVersion {

	result := CreateApplicationWithVersion(app, version)
	if rulesetBindings != nil {
		ary := make([]ApplicationApprovalRulesetBindingWithLatestApprovedVersion, 0, len(*rulesetBindings))
		for _, binding := range *rulesetBindings {
			ary = append(ary, CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(binding, binding.Version, false, true))
		}
		result.ApplicationBase.ApprovalRulesetBindings = &ary
	}
	return result
}

//
// ******** Constructor functions: WithLatestApprovedVersion ********
//

func CreateApplicationWithLatestApprovedVersion(app dbmodels.Application, version *dbmodels.ApplicationVersion) ApplicationWithLatestApprovedVersion {
	result := ApplicationWithLatestApprovedVersion{
		ReviewableBase: createReviewableBase(app.ReviewableBase),
		ApplicationBase: ApplicationBase{
			ID: app.ID,
		},
	}
	if version != nil {
		jsonStruct := CreateApplicationVersion(*version)
		result.LatestApprovedVersion = &jsonStruct
	}
	return result
}

func CreateApplicationWithLatestApprovedVersionAndRulesetBindings(app dbmodels.Application, version *dbmodels.ApplicationVersion, bindings []dbmodels.ApplicationApprovalRulesetBinding) ApplicationWithLatestApprovedVersion {
	bindingsJSON := make([]ApplicationApprovalRulesetBindingWithLatestApprovedVersion, 0, len(bindings))
	for _, binding := range bindings {
		bindingJSON := CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(binding, binding.Version, false, true)
		bindingsJSON = append(bindingsJSON, bindingJSON)
	}

	result := CreateApplicationWithLatestApprovedVersion(app, version)
	result.ApplicationBase.ApprovalRulesetBindings = &bindingsJSON
	return result
}

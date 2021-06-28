package json

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

//
// ******** Types, constants & variables ********
//

type ApprovalRulesetBase struct {
	ID                                 string                                                         `json:"id"`
	ApplicationApprovalRulesetBindings *[]ApplicationApprovalRulesetBindingWithApplicationAssociation `json:"application_approval_ruleset_bindings,omitempty"`
	NumBoundApplications               *uint                                                          `json:"num_bound_applications,omitempty"`
}

type ApprovalRulesetWithVersion struct {
	ReviewableBase
	ApprovalRulesetBase
	Version ApprovalRulesetVersion `json:"version"`
}

type ApprovalRulesetWithLatestApprovedVersion struct {
	ReviewableBase
	ApprovalRulesetBase
	LatestApprovedVersion *ApprovalRulesetVersion `json:"latest_approved_version"`
}

type ApprovalRulesetVersion struct {
	ReviewableVersionBase
	DisplayName        string          `json:"display_name"`
	Description        string          `json:"description"`
	GloballyApplicable bool            `json:"globally_applicable"`
	Enabled            bool            `json:"enabled"`
	ApprovalRules      *[]ApprovalRule `json:"approval_rules,omitempty"`

	NumBoundReleases               *uint                                                  `json:"num_bound_releases,omitempty"`
	ReleaseApprovalRulesetBindings *[]ReleaseApprovalRulesetBindingWithReleaseAssociation `json:"release_approval_ruleset_bindings,omitempty"`
}

//
// ******** ApprovalRulesetBase methods ********
//

func (base *ApprovalRulesetBase) PopulateFromDbmodelsApplicationApprovalRulesetBinding(bindings []dbmodels.ApplicationApprovalRulesetBinding) {
	bindingsJSON := make([]ApplicationApprovalRulesetBindingWithApplicationAssociation, 0, len(bindings))

	for _, binding := range bindings {
		if binding.Version == nil {
			panic("Application approval rule binding must have an associated version")
		}
		if binding.Version.VersionNumber == nil {
			panic("Application approval rule binding's version must be finalized")
		}
		if binding.Version.Adjustment == nil {
			panic("Application approval rule binding must have an associated adjustment")
		}

		bindingsJSON = append(bindingsJSON,
			CreateFromDbApplicationApprovalRulesetBindingWithApplicationAssociation(binding,
				*binding.Version, *binding.Version.Adjustment))
	}

	base.ApplicationApprovalRulesetBindings = &bindingsJSON
}

//
// ******** ApprovalRulesetVersion methods ********
//

func (version *ApprovalRulesetVersion) PopulateFromDbmodelsReleaseApprovalRulesetBindings(bindings []dbmodels.ReleaseApprovalRulesetBinding) {
	bindingsJSON := make([]ReleaseApprovalRulesetBindingWithReleaseAssociation, 0, len(bindings))
	for _, binding := range bindings {
		bindingsJSON = append(bindingsJSON,
			CreateFromDbReleaseApprovalRulesetBindingWithReleaseAssociation(binding))
	}
	version.ReleaseApprovalRulesetBindings = &bindingsJSON
}

func (version *ApprovalRulesetVersion) PopulateFromDbmodelsApprovalRulesetContents(contents dbmodels.ApprovalRulesetContents) {
	var ruleTypesProcessed uint = 0

	version.ApprovalRules = &[]ApprovalRule{}

	ruleTypesProcessed++
	for _, rule := range contents.HTTPApiApprovalRules {
		ruleJSON := ApprovalRule{Type: dbmodels.HTTPApiApprovalRuleType, HTTPApiApprovalRule: rule}
		*version.ApprovalRules = append(*version.ApprovalRules, ruleJSON)
	}

	ruleTypesProcessed++
	for _, rule := range contents.ScheduleApprovalRules {
		ruleJSON := ApprovalRule{Type: dbmodels.ScheduleApprovalRuleType, ScheduleApprovalRule: rule}
		*version.ApprovalRules = append(*version.ApprovalRules, ruleJSON)
	}

	ruleTypesProcessed++
	for _, rule := range contents.ManualApprovalRules {
		ruleJSON := ApprovalRule{Type: dbmodels.ManualApprovalRuleType, ManualApprovalRule: rule}
		*version.ApprovalRules = append(*version.ApprovalRules, ruleJSON)
	}

	if ruleTypesProcessed != dbmodels.NumApprovalRuleTypes {
		panic("Bug: code does not cover all approval rule types")
	}
}

//
// ******** Constructor functions ********
//

func CreateApprovalRulesetVersion(version dbmodels.ApprovalRulesetVersion, latestAdjustment dbmodels.ApprovalRulesetAdjustment) ApprovalRulesetVersion {
	return ApprovalRulesetVersion{
		ReviewableVersionBase: createReviewableVersionBase(version.ReviewableVersionBase, latestAdjustment.ReviewableAdjustmentBase),
		DisplayName:           latestAdjustment.DisplayName,
		Description:           latestAdjustment.Description,
		GloballyApplicable:    latestAdjustment.GloballyApplicable,
		Enabled:               latestAdjustment.Enabled,
	}
}

func CreateApprovalRulesetVersionWithStatsAndRules(version dbmodels.ApprovalRulesetVersion, latestAdjustment dbmodels.ApprovalRulesetAdjustment) ApprovalRulesetVersion {
	versionJSON := CreateApprovalRulesetVersion(version, latestAdjustment)
	versionJSON.NumBoundReleases = &latestAdjustment.NumBoundReleases
	versionJSON.PopulateFromDbmodelsApprovalRulesetContents(latestAdjustment.Rules)
	return versionJSON
}

func CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset dbmodels.ApprovalRuleset, version dbmodels.ApprovalRulesetVersion, latestAdjustment dbmodels.ApprovalRulesetAdjustment,
	appBindings []dbmodels.ApplicationApprovalRulesetBinding, releaseBindings []dbmodels.ReleaseApprovalRulesetBinding, rules dbmodels.ApprovalRulesetContents) ApprovalRulesetWithVersion {

	result := ApprovalRulesetWithVersion{
		ReviewableBase: createReviewableBase(ruleset.ReviewableBase),
		ApprovalRulesetBase: ApprovalRulesetBase{
			ID: ruleset.ID,
		},
		Version: CreateApprovalRulesetVersion(version, latestAdjustment),
	}

	result.ApprovalRulesetBase.PopulateFromDbmodelsApplicationApprovalRulesetBinding(appBindings)
	result.Version.PopulateFromDbmodelsReleaseApprovalRulesetBindings(releaseBindings)
	result.Version.PopulateFromDbmodelsApprovalRulesetContents(rules)

	return result
}

func CreateApprovalRulesetWithLatestApprovedVersion(ruleset dbmodels.ApprovalRuleset, version dbmodels.ApprovalRulesetVersion, latestAdjustment dbmodels.ApprovalRulesetAdjustment) ApprovalRulesetWithLatestApprovedVersion {
	versionJSON := CreateApprovalRulesetVersion(version, latestAdjustment)

	return ApprovalRulesetWithLatestApprovedVersion{
		ReviewableBase: createReviewableBase(ruleset.ReviewableBase),
		ApprovalRulesetBase: ApprovalRulesetBase{
			ID: ruleset.ID,
		},
		LatestApprovedVersion: &versionJSON,
	}
}

func CreateApprovalRulesetWithLatestApprovedVersionAndStats(ruleset dbmodels.ApprovalRulesetWithStats, version dbmodels.ApprovalRulesetVersion, latestAdjustment dbmodels.ApprovalRulesetAdjustment) ApprovalRulesetWithLatestApprovedVersion {
	versionJSON := CreateApprovalRulesetVersion(version, latestAdjustment)
	versionJSON.NumBoundReleases = &ruleset.NumBoundReleases

	return ApprovalRulesetWithLatestApprovedVersion{
		ReviewableBase: createReviewableBase(ruleset.ReviewableBase),
		ApprovalRulesetBase: ApprovalRulesetBase{
			ID:                   ruleset.ID,
			NumBoundApplications: &ruleset.NumBoundApplications,
		},
		LatestApprovedVersion: &versionJSON,
	}
}

func CreateApprovalRulesetWithLatestApprovedVersionAndBindingsAndRules(ruleset dbmodels.ApprovalRuleset, version dbmodels.ApprovalRulesetVersion, latestAdjustment dbmodels.ApprovalRulesetAdjustment,
	appBindings []dbmodels.ApplicationApprovalRulesetBinding, releaseBindings []dbmodels.ReleaseApprovalRulesetBinding, rules dbmodels.ApprovalRulesetContents) ApprovalRulesetWithLatestApprovedVersion {

	versionJSON := CreateApprovalRulesetVersion(version, latestAdjustment)
	versionJSON.PopulateFromDbmodelsReleaseApprovalRulesetBindings(releaseBindings)
	versionJSON.PopulateFromDbmodelsApprovalRulesetContents(rules)

	result := ApprovalRulesetWithLatestApprovedVersion{
		ReviewableBase: createReviewableBase(ruleset.ReviewableBase),
		ApprovalRulesetBase: ApprovalRulesetBase{
			ID: ruleset.ID,
		},
		LatestApprovedVersion: &versionJSON,
	}
	result.ApprovalRulesetBase.PopulateFromDbmodelsApplicationApprovalRulesetBinding(appBindings)

	return result
}

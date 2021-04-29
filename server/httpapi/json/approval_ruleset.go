package json

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type ApprovalRuleset struct {
	ID                 string    `json:"id"`
	MajorVersionNumber *uint32   `json:"major_version_number"`
	MinorVersionNumber uint32    `json:"minor_version_number"`
	DisplayName        string    `json:"display_name"`
	Description        string    `json:"description"`
	GloballyApplicable bool      `json:"globally_applicable"`
	ReviewState        string    `json:"review_state"`
	ReviewComments     *string   `json:"review_comments"`
	Enabled            bool      `json:"enabled"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ApprovalRulesetWithStats struct {
	ApprovalRuleset
	NumBoundApplications uint `json:"num_bound_applications"`
	NumBoundReleases     uint `json:"num_bound_releases"`
}

type ApprovalRulesetWithBindingAndRuleAssocations struct {
	ApprovalRuleset
	ApplicationApprovalRulesetBindings []ApplicationApprovalRulesetBindingWithApplicationAssociation `json:"application_approval_ruleset_bindings"`
	ReleaseApprovalRulesetBindings     []ReleaseApprovalRulesetBindingWithReleaseAssociation         `json:"release_approval_ruleset_bindings"`
	ApprovalRules                      []map[string]interface{}                                      `json:"approval_rules"`
}

func CreateFromDbApprovalRuleset(ruleset dbmodels.ApprovalRuleset, majorVersion dbmodels.ApprovalRulesetMajorVersion, minorVersion dbmodels.ApprovalRulesetMinorVersion) ApprovalRuleset {
	var reviewComments *string
	if minorVersion.ReviewComments.Valid {
		reviewComments = &minorVersion.ReviewComments.String
	}

	result := ApprovalRuleset{
		ID:                 ruleset.ID,
		MajorVersionNumber: majorVersion.VersionNumber,
		MinorVersionNumber: minorVersion.VersionNumber,
		DisplayName:        minorVersion.DisplayName,
		Description:        minorVersion.Description,
		GloballyApplicable: minorVersion.GloballyApplicable,
		ReviewState:        string(minorVersion.ReviewState),
		ReviewComments:     reviewComments,
		Enabled:            minorVersion.Enabled,
		CreatedAt:          ruleset.CreatedAt,
		UpdatedAt:          minorVersion.CreatedAt,
	}
	return result
}

func CreateFromDbApprovalRulesetWithStats(ruleset dbmodels.ApprovalRulesetWithStats, majorVersion dbmodels.ApprovalRulesetMajorVersion,
	minorVersion dbmodels.ApprovalRulesetMinorVersion) ApprovalRulesetWithStats {

	result := ApprovalRulesetWithStats{
		ApprovalRuleset:      CreateFromDbApprovalRuleset(ruleset.ApprovalRuleset, majorVersion, minorVersion),
		NumBoundApplications: ruleset.NumBoundApplications,
		NumBoundReleases:     ruleset.NumBoundReleases,
	}
	return result
}

func CreateFromDbApprovalRulesetWithBindingAndRuleAssociations(ruleset dbmodels.ApprovalRuleset, majorVersion dbmodels.ApprovalRulesetMajorVersion, minorVersion dbmodels.ApprovalRulesetMinorVersion,
	appBindings []dbmodels.ApplicationApprovalRulesetBinding, releaseBindings []dbmodels.ReleaseApprovalRulesetBinding, rules dbmodels.ApprovalRulesetContents) ApprovalRulesetWithBindingAndRuleAssocations {

	var ruleTypesProcessed uint = 0

	result := ApprovalRulesetWithBindingAndRuleAssocations{
		ApprovalRuleset:                    CreateFromDbApprovalRuleset(ruleset, majorVersion, minorVersion),
		ApplicationApprovalRulesetBindings: make([]ApplicationApprovalRulesetBindingWithApplicationAssociation, 0, len(appBindings)),
		ReleaseApprovalRulesetBindings:     make([]ReleaseApprovalRulesetBindingWithReleaseAssociation, 0, len(releaseBindings)),
		ApprovalRules:                      make([]map[string]interface{}, 0),
	}

	for _, binding := range appBindings {
		if binding.LatestMajorVersion == nil {
			panic("Application approval rule binding must have an associated latest major version")
		}
		if binding.LatestMajorVersion.VersionNumber == nil {
			panic("Application approval rule binding's latest major version must be finalized")
		}
		if binding.LatestMinorVersion == nil {
			panic("Application approval rule binding must have an associated latest minor version")
		}

		result.ApplicationApprovalRulesetBindings = append(result.ApplicationApprovalRulesetBindings,
			CreateFromDbApplicationApprovalRulesetBindingWithApplicationAssociation(binding,
				*binding.LatestMajorVersion, *binding.LatestMinorVersion))
	}

	for _, binding := range releaseBindings {
		result.ReleaseApprovalRulesetBindings = append(result.ReleaseApprovalRulesetBindings,
			CreateFromDbReleaseApprovalRulesetBindingWithReleaseAssociation(binding))
	}

	ruleTypesProcessed++
	for _, rule := range rules.HTTPApiApprovalRules {
		subJSON := CreateFromDbApprovalRule(rule.ApprovalRule)
		subJSON["type"] = "http_api"
		// TODO
		result.ApprovalRules = append(result.ApprovalRules, subJSON)
	}

	ruleTypesProcessed++
	for _, rule := range rules.ScheduleApprovalRules {
		subJSON := CreateFromDbApprovalRule(rule.ApprovalRule)
		subJSON["type"] = "schedule"
		subJSON["begin_time"] = getSqlStringContentsOrNil(rule.BeginTime)
		subJSON["end_time"] = getSqlStringContentsOrNil(rule.EndTime)
		subJSON["days_of_week"] = getSqlStringContentsOrNil(rule.DaysOfWeek)
		subJSON["days_of_month"] = getSqlStringContentsOrNil(rule.DaysOfMonth)
		subJSON["months_of_year"] = getSqlStringContentsOrNil(rule.MonthsOfYear)
		result.ApprovalRules = append(result.ApprovalRules, subJSON)
	}

	ruleTypesProcessed++
	for _, rule := range rules.ManualApprovalRules {
		subJSON := CreateFromDbApprovalRule(rule.ApprovalRule)
		subJSON["type"] = "manual"
		// TODO
		result.ApprovalRules = append(result.ApprovalRules, subJSON)
	}

	if ruleTypesProcessed != dbmodels.NumApprovalRuleTypes {
		panic("Bug: code does not cover all approval rule types")
	}

	return result
}

func CreateFromDbApprovalRule(rule dbmodels.ApprovalRule) map[string]interface{} {
	return map[string]interface{}{
		"id":         rule.ID,
		"enabled":    rule.Enabled,
		"created_at": rule.CreatedAt,
	}
}

func getSqlStringContentsOrNil(str sql.NullString) interface{} {
	if str.Valid {
		return str.String
	}
	return nil
}

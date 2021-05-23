package json

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

type ApprovalRuleset struct {
	ID                 string    `json:"id"`
	VersionNumber      *uint32   `json:"version_number"`
	AdjustmentNumber   uint32    `json:"adjustment_number"`
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

func CreateFromDbApprovalRuleset(ruleset dbmodels.ApprovalRuleset, version dbmodels.ApprovalRulesetVersion, adjustment dbmodels.ApprovalRulesetAdjustment) ApprovalRuleset {
	var reviewComments *string
	if adjustment.ReviewComments.Valid {
		reviewComments = &adjustment.ReviewComments.String
	}

	result := ApprovalRuleset{
		ID:                 ruleset.ID,
		VersionNumber:      version.VersionNumber,
		AdjustmentNumber:   adjustment.AdjustmentNumber,
		DisplayName:        adjustment.DisplayName,
		Description:        adjustment.Description,
		GloballyApplicable: adjustment.GloballyApplicable,
		ReviewState:        string(adjustment.ReviewState),
		ReviewComments:     reviewComments,
		Enabled:            adjustment.Enabled,
		CreatedAt:          ruleset.CreatedAt,
		UpdatedAt:          adjustment.CreatedAt,
	}
	return result
}

func CreateFromDbApprovalRulesetWithStats(ruleset dbmodels.ApprovalRulesetWithStats, version dbmodels.ApprovalRulesetVersion,
	adjustment dbmodels.ApprovalRulesetAdjustment) ApprovalRulesetWithStats {

	result := ApprovalRulesetWithStats{
		ApprovalRuleset:      CreateFromDbApprovalRuleset(ruleset.ApprovalRuleset, version, adjustment),
		NumBoundApplications: ruleset.NumBoundApplications,
		NumBoundReleases:     ruleset.NumBoundReleases,
	}
	return result
}

func CreateFromDbApprovalRulesetWithBindingAndRuleAssociations(ruleset dbmodels.ApprovalRuleset, version dbmodels.ApprovalRulesetVersion, adjustment dbmodels.ApprovalRulesetAdjustment,
	appBindings []dbmodels.ApplicationApprovalRulesetBinding, releaseBindings []dbmodels.ReleaseApprovalRulesetBinding, rules dbmodels.ApprovalRulesetContents) ApprovalRulesetWithBindingAndRuleAssocations {

	var ruleTypesProcessed uint = 0

	result := ApprovalRulesetWithBindingAndRuleAssocations{
		ApprovalRuleset:                    CreateFromDbApprovalRuleset(ruleset, version, adjustment),
		ApplicationApprovalRulesetBindings: make([]ApplicationApprovalRulesetBindingWithApplicationAssociation, 0, len(appBindings)),
		ReleaseApprovalRulesetBindings:     make([]ReleaseApprovalRulesetBindingWithReleaseAssociation, 0, len(releaseBindings)),
		ApprovalRules:                      make([]map[string]interface{}, 0),
	}

	for _, binding := range appBindings {
		if binding.LatestVersion == nil {
			panic("Application approval rule binding must have an associated latest version")
		}
		if binding.LatestVersion.VersionNumber == nil {
			panic("Application approval rule binding's latest version must be finalized")
		}
		if binding.LatestAdjustment == nil {
			panic("Application approval rule binding must have an associated latest adjustment")
		}

		result.ApplicationApprovalRulesetBindings = append(result.ApplicationApprovalRulesetBindings,
			CreateFromDbApplicationApprovalRulesetBindingWithApplicationAssociation(binding,
				*binding.LatestVersion, *binding.LatestAdjustment))
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

package dbmodels

import "github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"

func (ruleset ApprovalRuleset) GetPrimaryKey() interface{} {
	return ruleset.ID
}

func (ruleset *ApprovalRuleset) SetLatestVersion(version IReviewableVersion) {
	ruleset.LatestVersion = version.(*ApprovalRulesetVersion)
}

func (ruleset *ApprovalRuleset) SetLatestAdjustment(adjustment IReviewableAdjustment) {
	ruleset.LatestAdjustment = adjustment.(*ApprovalRulesetAdjustment)
}

// NewDraftVersion returns an unsaved ApprovalRulesetVersion and ApprovalRulesetAdjustment
// in draft proposal state.
func (ruleset ApprovalRuleset) NewDraftVersion() (*ApprovalRulesetVersion, *ApprovalRulesetAdjustment) {
	var adjustment ApprovalRulesetAdjustment
	var version *ApprovalRulesetVersion = &adjustment.ApprovalRulesetVersion

	if ruleset.LatestAdjustment != nil {
		adjustment = *ruleset.LatestAdjustment
	}

	version.BaseModel = ruleset.BaseModel
	version.ReviewableVersionBase = ReviewableVersionBase{}
	version.ApprovalRuleset = ruleset
	version.ApprovalRulesetID = ruleset.ID

	adjustment.BaseModel = ruleset.BaseModel
	adjustment.ApprovalRulesetVersionID = 0
	adjustment.ReviewableAdjustmentBase = ReviewableAdjustmentBase{
		AdjustmentNumber: 1,
		ReviewState:      reviewstate.Draft,
	}

	return version, &adjustment
}

func (version ApprovalRulesetVersion) GetID() interface{} {
	return version.ID
}

func (version ApprovalRulesetVersion) GetReviewablePrimaryKey() interface{} {
	return version.ApprovalRulesetID
}

func (version *ApprovalRulesetVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApprovalRuleset)
	version.ApprovalRulesetID = ruleset.ID
	version.ApprovalRuleset = *ruleset
}

func (adjustment ApprovalRulesetAdjustment) GetVersionID() interface{} {
	return adjustment.ApprovalRulesetVersionID
}

func (adjustment *ApprovalRulesetAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApprovalRulesetVersion)
	adjustment.ApprovalRulesetVersionID = concreteVersion.ID
	adjustment.ApprovalRulesetVersion = *concreteVersion
}

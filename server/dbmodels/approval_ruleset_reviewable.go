package dbmodels

import "github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"

func (ruleset ApprovalRuleset) GetPrimaryKey() interface{} {
	return ruleset.ID
}

func (ruleset ApprovalRuleset) GetPrimaryKeyGormValue() []interface{} {
	return []interface{}{ruleset.ID}
}

func (ruleset *ApprovalRuleset) AssociateWithVersion(version IReviewableVersion) {
	ruleset.Version = version.(*ApprovalRulesetVersion)
}

// NewDraftVersion returns an unsaved ApprovalRulesetVersion and ApprovalRulesetAdjustment
// in draft proposal state.
func (ruleset ApprovalRuleset) NewDraftVersion() (*ApprovalRulesetVersion, *ApprovalRulesetAdjustment) {
	var adjustment ApprovalRulesetAdjustment
	var version *ApprovalRulesetVersion = &adjustment.ApprovalRulesetVersion

	if ruleset.Version != nil && ruleset.Version.Adjustment != nil {
		adjustment = *ruleset.Version.Adjustment
	}

	version.BaseModel = ruleset.BaseModel
	version.ReviewableVersionBase = ReviewableVersionBase{}
	version.ApprovalRuleset = ruleset
	version.ApprovalRulesetID = ruleset.ID
	version.Adjustment = &adjustment

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

func (version ApprovalRulesetVersion) GetReviewablePrimaryKeyGormValue() []interface{} {
	return []interface{}{version.ApprovalRulesetID}
}

func (version *ApprovalRulesetVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApprovalRuleset)
	version.ApprovalRulesetID = ruleset.ID
	version.ApprovalRuleset = *ruleset
}

func (version *ApprovalRulesetVersion) AssociateWithAdjustment(adjustment IReviewableAdjustment) {
	version.Adjustment = adjustment.(*ApprovalRulesetAdjustment)
}

func (adjustment ApprovalRulesetAdjustment) GetVersionID() interface{} {
	return adjustment.ApprovalRulesetVersionID
}

func (adjustment *ApprovalRulesetAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApprovalRulesetVersion)
	adjustment.ApprovalRulesetVersionID = concreteVersion.ID
	adjustment.ApprovalRulesetVersion = *concreteVersion
}

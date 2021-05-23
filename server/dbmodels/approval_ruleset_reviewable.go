package dbmodels

func (ruleset ApprovalRuleset) GetPrimaryKey() interface{} {
	return ruleset.ID
}

func (ruleset *ApprovalRuleset) SetLatestVersion(version IReviewableVersion) {
	ruleset.LatestVersion = version.(*ApprovalRulesetVersion)
}

func (ruleset *ApprovalRuleset) SetLatestAdjustment(adjustment IReviewableAdjustment) {
	ruleset.LatestAdjustment = adjustment.(*ApprovalRulesetAdjustment)
}

func (major ApprovalRulesetVersion) GetID() interface{} {
	return major.ID
}

func (major ApprovalRulesetVersion) GetReviewablePrimaryKey() interface{} {
	return major.ApprovalRulesetID
}

func (major *ApprovalRulesetVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApprovalRuleset)
	major.ApprovalRulesetID = ruleset.ID
	major.ApprovalRuleset = *ruleset
}

func (adjustment ApprovalRulesetAdjustment) GetVersionID() interface{} {
	return adjustment.ApprovalRulesetVersionID
}

func (adjustment *ApprovalRulesetAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApprovalRulesetVersion)
	adjustment.ApprovalRulesetVersionID = concreteVersion.ID
	adjustment.ApprovalRulesetVersion = *concreteVersion
}

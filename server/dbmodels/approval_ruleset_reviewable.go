package dbmodels

func (ruleset ApprovalRuleset) GetPrimaryKey() interface{} {
	return ruleset.ID
}

func (ruleset ApprovalRuleset) GetPrimaryKeyGormValue() []interface{} {
	return []interface{}{ruleset.ID}
}

func (ruleset *ApprovalRuleset) AssociateWithVersion(version IReviewableVersion) {
	ruleset.Version = version.(*ApprovalRulesetVersion)
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

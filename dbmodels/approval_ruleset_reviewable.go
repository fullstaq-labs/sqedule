package dbmodels

// GetPrimaryKey ...
func (ruleset ApprovalRuleset) GetPrimaryKey() interface{} {
	return ruleset.ID
}

// SetLatestMajorVersion ...
func (ruleset *ApprovalRuleset) SetLatestMajorVersion(majorVersion IReviewableMajorVersion) {
	ruleset.LatestMajorVersion = majorVersion.(*ApprovalRulesetMajorVersion)
}

// SetLatestMinorVersion ...
func (ruleset *ApprovalRuleset) SetLatestMinorVersion(minorVersion IReviewableMinorVersion) {
	ruleset.LatestMinorVersion = minorVersion.(*ApprovalRulesetMinorVersion)
}

// GetID ...
func (major ApprovalRulesetMajorVersion) GetID() interface{} {
	return major.ID
}

// GetReviewablePrimaryKey ...
func (major ApprovalRulesetMajorVersion) GetReviewablePrimaryKey() interface{} {
	return major.ApprovalRulesetID
}

// AssociateWithReviewable ...
func (major *ApprovalRulesetMajorVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApprovalRuleset)
	major.ApprovalRulesetID = ruleset.ID
	major.ApprovalRuleset = *ruleset
}

// GetMajorVersionID ...
func (minor ApprovalRulesetMinorVersion) GetMajorVersionID() interface{} {
	return minor.ApprovalRulesetMajorVersionID
}

// AssociateWithMajorVersion ...
func (minor *ApprovalRulesetMinorVersion) AssociateWithMajorVersion(majorVersion IReviewableMajorVersion) {
	concreteMajorVersion := majorVersion.(*ApprovalRulesetMajorVersion)
	minor.ApprovalRulesetMajorVersionID = concreteMajorVersion.ID
	minor.ApprovalRulesetMajorVersion = *concreteMajorVersion
}

package dbmodels

// GormValue ...
func (primaryKey ApplicationApprovalRulesetBindingPrimaryKey) GormValue() interface{} {
	return []interface{}{primaryKey.ApplicationID, primaryKey.ApprovalRulesetID}
}

// GetPrimaryKey ...
func (binding ApplicationApprovalRulesetBinding) GetPrimaryKey() interface{} {
	return binding.ApplicationApprovalRulesetBindingPrimaryKey
}

// SetLatestMajorVersion ...
func (binding *ApplicationApprovalRulesetBinding) SetLatestMajorVersion(majorVersion IReviewableMajorVersion) {
	binding.LatestMajorVersion = majorVersion.(*ApplicationApprovalRulesetBindingMajorVersion)
}

// SetLatestMinorVersion ...
func (binding *ApplicationApprovalRulesetBinding) SetLatestMinorVersion(minorVersion IReviewableMinorVersion) {
	binding.LatestMinorVersion = minorVersion.(*ApplicationApprovalRulesetBindingMinorVersion)
}

// GetID ...
func (major ApplicationApprovalRulesetBindingMajorVersion) GetID() interface{} {
	return major.ID
}

// GetReviewablePrimaryKey ...
func (major ApplicationApprovalRulesetBindingMajorVersion) GetReviewablePrimaryKey() interface{} {
	return ApplicationApprovalRulesetBindingPrimaryKey{
		ApplicationID:     major.ApplicationID,
		ApprovalRulesetID: major.ApprovalRulesetID,
	}
}

// AssociateWithReviewable ...
func (major *ApplicationApprovalRulesetBindingMajorVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApplicationApprovalRulesetBinding)
	major.ApplicationID = ruleset.ApplicationID
	major.ApprovalRulesetID = ruleset.ApprovalRulesetID
	major.ApplicationApprovalRulesetBinding = *ruleset
}

// GetMajorVersionID ...
func (minor ApplicationApprovalRulesetBindingMinorVersion) GetMajorVersionID() interface{} {
	return minor.ApplicationApprovalRulesetBindingMajorVersionID
}

// AssociateWithMajorVersion ...
func (minor *ApplicationApprovalRulesetBindingMinorVersion) AssociateWithMajorVersion(majorVersion IReviewableMajorVersion) {
	concreteMajorVersion := majorVersion.(*ApplicationApprovalRulesetBindingMajorVersion)
	minor.ApplicationApprovalRulesetBindingMajorVersionID = concreteMajorVersion.ID
	minor.ApplicationApprovalRulesetBindingMajorVersion = *concreteMajorVersion
}

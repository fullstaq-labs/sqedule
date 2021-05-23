package dbmodels

func (primaryKey ApplicationApprovalRulesetBindingPrimaryKey) GormValue() interface{} {
	return []interface{}{primaryKey.ApplicationID, primaryKey.ApprovalRulesetID}
}

func (binding ApplicationApprovalRulesetBinding) GetPrimaryKey() interface{} {
	return binding.ApplicationApprovalRulesetBindingPrimaryKey
}

func (binding *ApplicationApprovalRulesetBinding) SetLatestVersion(version IReviewableVersion) {
	binding.LatestVersion = version.(*ApplicationApprovalRulesetBindingVersion)
}

func (binding *ApplicationApprovalRulesetBinding) SetLatestAdjustment(adjustment IReviewableAdjustment) {
	binding.LatestAdjustment = adjustment.(*ApplicationApprovalRulesetBindingAdjustment)
}

func (major ApplicationApprovalRulesetBindingVersion) GetID() interface{} {
	return major.ID
}

func (major ApplicationApprovalRulesetBindingVersion) GetReviewablePrimaryKey() interface{} {
	return ApplicationApprovalRulesetBindingPrimaryKey{
		ApplicationID:     major.ApplicationID,
		ApprovalRulesetID: major.ApprovalRulesetID,
	}
}

func (major *ApplicationApprovalRulesetBindingVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApplicationApprovalRulesetBinding)
	major.ApplicationID = ruleset.ApplicationID
	major.ApprovalRulesetID = ruleset.ApprovalRulesetID
	major.ApplicationApprovalRulesetBinding = *ruleset
}

func (adjustment ApplicationApprovalRulesetBindingAdjustment) GetVersionID() interface{} {
	return adjustment.ApplicationApprovalRulesetBindingVersionID
}

func (adjustment *ApplicationApprovalRulesetBindingAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApplicationApprovalRulesetBindingVersion)
	adjustment.ApplicationApprovalRulesetBindingVersionID = concreteVersion.ID
	adjustment.ApplicationApprovalRulesetBindingVersion = *concreteVersion
}

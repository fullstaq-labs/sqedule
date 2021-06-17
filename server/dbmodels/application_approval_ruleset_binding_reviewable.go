package dbmodels

func (binding ApplicationApprovalRulesetBinding) GetPrimaryKey() interface{} {
	return binding.ApplicationApprovalRulesetBindingPrimaryKey
}

func (binding ApplicationApprovalRulesetBinding) GetPrimaryKeyGormValue() []interface{} {
	return []interface{}{binding.ApplicationID, binding.ApprovalRulesetID}
}

func (binding *ApplicationApprovalRulesetBinding) AssociateWithVersion(version IReviewableVersion) {
	binding.Version = version.(*ApplicationApprovalRulesetBindingVersion)
}

func (version ApplicationApprovalRulesetBindingVersion) GetID() interface{} {
	return version.ID
}

func (version ApplicationApprovalRulesetBindingVersion) GetReviewablePrimaryKey() interface{} {
	return ApplicationApprovalRulesetBindingPrimaryKey{
		ApplicationID:     version.ApplicationID,
		ApprovalRulesetID: version.ApprovalRulesetID,
	}
}

func (version *ApplicationApprovalRulesetBindingVersion) AssociateWithReviewable(reviewable IReviewable) {
	ruleset := reviewable.(*ApplicationApprovalRulesetBinding)
	version.ApplicationID = ruleset.ApplicationID
	version.ApprovalRulesetID = ruleset.ApprovalRulesetID
	version.ApplicationApprovalRulesetBinding = *ruleset
}

func (version *ApplicationApprovalRulesetBindingVersion) AssociateWithAdjustment(adjustment IReviewableAdjustment) {
	version.Adjustment = adjustment.(*ApplicationApprovalRulesetBindingAdjustment)
}

func (adjustment ApplicationApprovalRulesetBindingAdjustment) GetVersionID() interface{} {
	return adjustment.ApplicationApprovalRulesetBindingVersionID
}

func (adjustment *ApplicationApprovalRulesetBindingAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApplicationApprovalRulesetBindingVersion)
	adjustment.ApplicationApprovalRulesetBindingVersionID = concreteVersion.ID
	adjustment.ApplicationApprovalRulesetBindingVersion = *concreteVersion
}

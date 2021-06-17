package dbmodels

func (app Application) GetPrimaryKey() interface{} {
	return app.ID
}

func (app Application) GetPrimaryKeyGormValue() []interface{} {
	return []interface{}{app.ID}
}

func (app *Application) AssociateWithVersion(version IReviewableVersion) {
	app.Version = version.(*ApplicationVersion)
}

func (version ApplicationVersion) GetID() interface{} {
	return version.ID
}

func (version ApplicationVersion) GetReviewablePrimaryKey() interface{} {
	return version.ApplicationID
}

func (version ApplicationVersion) GetReviewablePrimaryKeyGormValue() []interface{} {
	return []interface{}{version.ApplicationID}
}

func (version *ApplicationVersion) AssociateWithReviewable(reviewable IReviewable) {
	application := reviewable.(*Application)
	version.ApplicationID = application.ID
	version.Application = *application
}

func (version *ApplicationVersion) AssociateWithAdjustment(adjustment IReviewableAdjustment) {
	version.Adjustment = adjustment.(*ApplicationAdjustment)
}

func (adjustment ApplicationAdjustment) GetVersionID() interface{} {
	return adjustment.ApplicationVersionID
}

func (adjustment *ApplicationAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApplicationVersion)
	adjustment.ApplicationVersionID = concreteVersion.ID
	adjustment.ApplicationVersion = *concreteVersion
}

package dbmodels

// GetPrimaryKey ...
func (app Application) GetPrimaryKey() interface{} {
	return app.ID
}

func (app *Application) SetLatestVersion(version IReviewableVersion) {
	app.LatestVersion = version.(*ApplicationVersion)
}

func (app *Application) SetLatestAdjustment(adjustment IReviewableAdjustment) {
	app.LatestAdjustment = adjustment.(*ApplicationAdjustment)
}

func (major ApplicationVersion) GetID() interface{} {
	return major.ID
}

func (major ApplicationVersion) GetReviewablePrimaryKey() interface{} {
	return major.ApplicationID
}

func (major *ApplicationVersion) AssociateWithReviewable(reviewable IReviewable) {
	application := reviewable.(*Application)
	major.ApplicationID = application.ID
	major.Application = *application
}

func (adjustment ApplicationAdjustment) GetVersionID() interface{} {
	return adjustment.ApplicationVersionID
}

func (adjustment *ApplicationAdjustment) AssociateWithVersion(version IReviewableVersion) {
	concreteVersion := version.(*ApplicationVersion)
	adjustment.ApplicationVersionID = concreteVersion.ID
	adjustment.ApplicationVersion = *concreteVersion
}

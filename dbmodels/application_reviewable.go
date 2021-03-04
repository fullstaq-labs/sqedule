package dbmodels

// GetPrimaryKey ...
func (app Application) GetPrimaryKey() interface{} {
	return app.ID
}

// SetLatestMajorVersion ...
func (app *Application) SetLatestMajorVersion(majorVersion IReviewableMajorVersion) {
	app.LatestMajorVersion = majorVersion.(*ApplicationMajorVersion)
}

// SetLatestMinorVersion ...
func (app *Application) SetLatestMinorVersion(minorVersion IReviewableMinorVersion) {
	app.LatestMinorVersion = minorVersion.(*ApplicationMinorVersion)
}

// GetID ...
func (major ApplicationMajorVersion) GetID() interface{} {
	return major.ID
}

// GetReviewablePrimaryKey ...
func (major ApplicationMajorVersion) GetReviewablePrimaryKey() interface{} {
	return major.ApplicationID
}

// AssociateWithReviewable ...
func (major *ApplicationMajorVersion) AssociateWithReviewable(reviewable IReviewable) {
	application := reviewable.(*Application)
	major.ApplicationID = application.ID
	major.Application = *application
}

// GetMajorVersionID ...
func (minor ApplicationMinorVersion) GetMajorVersionID() interface{} {
	return minor.ApplicationMajorVersionID
}

// AssociateWithMajorVersion ...
func (minor *ApplicationMinorVersion) AssociateWithMajorVersion(majorVersion IReviewableMajorVersion) {
	concreteMajorVersion := majorVersion.(*ApplicationMajorVersion)
	minor.ApplicationMajorVersionID = concreteMajorVersion.ID
	minor.ApplicationMajorVersion = *concreteMajorVersion
}

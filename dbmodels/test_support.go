package dbmodels

import (
	"fmt"

	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/dbmodels/reviewstate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateMockOrganization ...
func CreateMockOrganization(db *gorm.DB) (Organization, error) {
	result := Organization{
		ID:          "org1",
		DisplayName: "Org 1",
	}
	tx := db.Create(&result)
	return result, tx.Error
}

// CreateMockApplicationWithOneVersion ...
func CreateMockApplicationWithOneVersion(db *gorm.DB, organization Organization) (Application, error) {
	result := Application{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ID: "app1",
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return Application{}, savetx.Error
	}

	var majorVersionNumber uint32 = 1
	var majorVersion = ApplicationMajorVersion{
		OrganizationID: organization.ID,
		Organization:   organization,
		ApplicationID:  result.ID,
		Application:    result,
		VersionNumber:  &majorVersionNumber,
	}
	savetx = db.Omit(clause.Associations).Create(&majorVersion)
	if savetx.Error != nil {
		return Application{}, savetx.Error
	}

	var minorVersion = ApplicationMinorVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationMajorVersionID: majorVersion.ID,
		ApplicationMajorVersion:   majorVersion,
		VersionNumber:             1,
		ReviewState:               reviewstate.Approved,
		DisplayName:               "App 1",
	}
	savetx = db.Omit(clause.Associations).Create(&minorVersion)
	if savetx.Error != nil {
		return Application{}, savetx.Error
	}

	result.LatestMajorVersion = &majorVersion
	result.LatestMinorVersion = &minorVersion
	return result, nil
}

// CreateMockReleaseWithInProgressState ...
func CreateMockReleaseWithInProgressState(db *gorm.DB, organization Organization, application Application,
	customizeFunc func(*Release)) (Release, error) {

	result := Release{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID: application.ID,
		Application:   application,
		State:         releasestate.InProgress,
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	tx := db.Omit(clause.Associations).Create(&result)
	return result, tx.Error
}

// CreateMockRulesetWith1Version ...
func CreateMockRulesetWith1Version(db *gorm.DB, organization Organization, id string, customizeFunc func(*ApprovalRulesetMinorVersion)) (ApprovalRuleset, error) {
	var savetx *gorm.DB
	var majorVersionNumber uint32 = 1

	ruleset := ApprovalRuleset{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ID: id,
	}
	savetx = db.Omit(clause.Associations).Create(&ruleset)
	if savetx.Error != nil {
		return ApprovalRuleset{}, fmt.Errorf("Error creating ApprovalRuleset: %w", savetx.Error)
	}

	majorVersion := ApprovalRulesetMajorVersion{
		OrganizationID:    organization.ID,
		Organization:      organization,
		ApprovalRulesetID: ruleset.ID,
		ApprovalRuleset:   ruleset,
		VersionNumber:     &majorVersionNumber,
	}
	savetx = db.Omit(clause.Associations).Create(&majorVersion)
	if savetx.Error != nil {
		return ApprovalRuleset{}, fmt.Errorf("Error creating ApprovalRulesetMajorVersion: %w", savetx.Error)
	}

	minorVersion := ApprovalRulesetMinorVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApprovalRulesetMajorVersionID: majorVersion.ID,
		ApprovalRulesetMajorVersion:   majorVersion,
		VersionNumber:                 1,
		ReviewState:                   reviewstate.Approved,
		DisplayName:                   "Ruleset",
		Description:                   "",
	}
	if customizeFunc != nil {
		customizeFunc(&minorVersion)
	}
	savetx = db.Omit(clause.Associations).Create(&minorVersion)
	if savetx.Error != nil {
		return ApprovalRuleset{}, fmt.Errorf("Error creating ApprovalRulesetMinorVersion: %w", savetx.Error)
	}

	ruleset.LatestMajorVersion = &majorVersion
	ruleset.LatestMinorVersion = &minorVersion
	return ruleset, nil
}

// CreateMockApplicationRulesetBindingWithEnforcingMode1Version ...
func CreateMockApplicationRulesetBindingWithEnforcingMode1Version(db *gorm.DB, organization Organization, application Application, ruleset ApprovalRuleset, customizeFunc func(*ApplicationApprovalRulesetBindingMinorVersion)) (ApplicationApprovalRulesetBinding, error) {
	var binding ApplicationApprovalRulesetBinding
	var majorVersion ApplicationApprovalRulesetBindingMajorVersion
	var minorVersion ApplicationApprovalRulesetBindingMinorVersion
	var err error

	binding = ApplicationApprovalRulesetBinding{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationApprovalRulesetBindingPrimaryKey: ApplicationApprovalRulesetBindingPrimaryKey{
			ApplicationID:     application.ID,
			ApprovalRulesetID: ruleset.ID,
		},
		Application:        application,
		ApprovalRuleset:    ruleset,
		LatestMajorVersion: &majorVersion,
		LatestMinorVersion: &minorVersion,
	}
	savetx := db.Omit(clause.Associations).Create(&binding)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBinding: %w", savetx.Error)
	}

	var versionNumber uint32 = 1
	majorVersion, err = CreateMockApplicationApprovalRulesetBindingMajorVersion(db, organization, application, binding, &versionNumber)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBindingMajorVersion: %w", err)
	}

	minorVersion, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(db, organization, majorVersion, customizeFunc)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBindingMinorVersion: %w", savetx.Error)
	}

	return binding, nil
}

// CreateMockApplicationApprovalRulesetBindingMajorVersion ...
func CreateMockApplicationApprovalRulesetBindingMajorVersion(db *gorm.DB, organization Organization, application Application, binding ApplicationApprovalRulesetBinding, versionNumber *uint32) (ApplicationApprovalRulesetBindingMajorVersion, error) {
	result := ApplicationApprovalRulesetBindingMajorVersion{
		OrganizationID:                    organization.ID,
		Organization:                      organization,
		ApplicationID:                     application.ID,
		ApprovalRulesetID:                 binding.ApprovalRulesetID,
		VersionNumber:                     versionNumber,
		ApplicationApprovalRulesetBinding: binding,
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBindingMajorVersion{}, savetx.Error
	}

	return result, nil
}

// CreateMockApplicationApprovalRulesetBindingMinorVersion ...
func CreateMockApplicationApprovalRulesetBindingMinorVersion(db *gorm.DB, organization Organization, majorVersion ApplicationApprovalRulesetBindingMajorVersion,
	customizeFunc func(*ApplicationApprovalRulesetBindingMinorVersion)) (ApplicationApprovalRulesetBindingMinorVersion, error) {

	result := ApplicationApprovalRulesetBindingMinorVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationApprovalRulesetBindingMajorVersionID: majorVersion.ID,
		ApplicationApprovalRulesetBindingMajorVersion:   majorVersion,
		VersionNumber: 1,
		ReviewState:   reviewstate.Approved,
		Enabled:       true,
		Mode:          approvalrulesetbindingmode.Enforcing,
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBindingMinorVersion{}, savetx.Error
	}

	return result, nil
}

// CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version creates two rulesets and bindings.
// One binding is in permissive mode, other in enforcing mode.
// Each binding, and each ruleset, has 1 version and is approved.
func CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(db *gorm.DB, organization Organization,
	application Application) (ApplicationApprovalRulesetBinding, ApplicationApprovalRulesetBinding, error) {

	// Ruleset 1: permissive

	ruleset1, err := CreateMockRulesetWith1Version(db, organization, "ruleset1", func(minorVersion *ApprovalRulesetMinorVersion) {
		minorVersion.DisplayName = "Ruleset 1"
	})
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, ApplicationApprovalRulesetBinding{}, err
	}
	ruleset1Binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(db, organization, application,
		ruleset1, func(minorVersion *ApplicationApprovalRulesetBindingMinorVersion) {
			minorVersion.Mode = approvalrulesetbindingmode.Permissive
		})
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, ApplicationApprovalRulesetBinding{}, err
	}

	// Ruleset 2: enforcing

	ruleset2, err := CreateMockRulesetWith1Version(db, organization, "ruleset2", func(minorVersion *ApprovalRulesetMinorVersion) {
		minorVersion.DisplayName = "Ruleset 2"
	})
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, ApplicationApprovalRulesetBinding{}, err
	}
	ruleset2Binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(db, organization, application,
		ruleset2, nil)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, ApplicationApprovalRulesetBinding{}, err
	}

	return ruleset1Binding, ruleset2Binding, nil
}

// CreateMockReleaseBackgroundJob ...
func CreateMockReleaseBackgroundJob(db *gorm.DB, organization Organization, app Application, release Release) (ReleaseBackgroundJob, error) {
	result := ReleaseBackgroundJob{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID: app.ID,
		LockID:        1,
		ReleaseID:     release.ID,
		Release:       release,
	}
	tx := db.Omit(clause.Associations).Create(&result)
	if tx.Error != nil {
		return ReleaseBackgroundJob{}, tx.Error
	}
	return result, nil
}

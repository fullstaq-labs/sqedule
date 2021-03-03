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

// CreateMockApprovalRulesetsAndBindingsWith2Modes1Version creates two rulesets and bindings.
// One binding is in permissive mode, other in enforcing mode.
// Each ruleset has 1 version and is approved.
func CreateMockApprovalRulesetsAndBindingsWith2Modes1Version(db *gorm.DB, organization Organization, application Application) (ApprovalRulesetBinding, ApprovalRulesetBinding, error) {
	var savetx *gorm.DB

	// Ruleset 1: permissive

	ruleset1, err := CreateMockRulesetWith1Version(db, organization, "ruleset1", func(minorVersion *ApprovalRulesetMinorVersion) {
		minorVersion.DisplayName = "Ruleset 1"
	})
	if err != nil {
		return ApprovalRulesetBinding{}, ApprovalRulesetBinding{}, err
	}

	ruleset1Binding := ApprovalRulesetBinding{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID:     application.ID,
		Application:       application,
		ApprovalRulesetID: ruleset1.ID,
		ApprovalRuleset:   ruleset1,
		Mode:              approvalrulesetbindingmode.Permissive,
	}
	savetx = db.Omit(clause.Associations).Create(&ruleset1Binding)
	if savetx.Error != nil {
		return ApprovalRulesetBinding{}, ApprovalRulesetBinding{}, savetx.Error
	}

	// Ruleset 2: enforcing

	ruleset2, err := CreateMockRulesetWith1Version(db, organization, "ruleset2", func(minorVersion *ApprovalRulesetMinorVersion) {
		minorVersion.DisplayName = "Ruleset 2"
	})
	if err != nil {
		return ApprovalRulesetBinding{}, ApprovalRulesetBinding{}, err
	}

	ruleset2Binding := ApprovalRulesetBinding{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID:     application.ID,
		Application:       application,
		ApprovalRulesetID: ruleset2.ID,
		ApprovalRuleset:   ruleset2,
		Mode:              approvalrulesetbindingmode.Enforcing,
	}
	savetx = db.Omit(clause.Associations).Create(&ruleset2Binding)
	if savetx.Error != nil {
		return ApprovalRulesetBinding{}, ApprovalRulesetBinding{}, savetx.Error
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

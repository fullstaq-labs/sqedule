package dbmodels

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/organizationmemberrole"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CreateMockOrganization ...
func CreateMockOrganization(db *gorm.DB, customizeFunc func(org *Organization)) (Organization, error) {
	result := Organization{
		ID:          "org1",
		DisplayName: "Org 1",
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	tx := db.Create(&result)
	return result, tx.Error
}

// CreateMockServiceAccountWithAdminRole ...
func CreateMockServiceAccountWithAdminRole(db *gorm.DB, organization Organization,
	customizeFunc func(sa *ServiceAccount)) (ServiceAccount, error) {

	result := ServiceAccount{
		OrganizationMember: OrganizationMember{
			BaseModel: BaseModel{
				OrganizationID: organization.ID,
				Organization:   organization,
			},
			Role:         organizationmemberrole.Admin,
			PasswordHash: "unauthenticatable",
		},
		Name: "sa1",
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return ServiceAccount{}, savetx.Error
	}
	return result, nil
}

// CreateMockApplicationWith1Version ...
func CreateMockApplicationWith1Version(db *gorm.DB, organization Organization, customizeFunc func(*Application), adjustmentCustomizeFunc func(*ApplicationAdjustment)) (Application, error) {
	result := Application{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ID: "app1",
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return Application{}, savetx.Error
	}

	var versionNumber uint32 = 1
	var version = ApplicationVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ReviewableVersionBase: ReviewableVersionBase{
			VersionNumber: &versionNumber,
			ApprovedAt:    sql.NullTime{Time: time.Now(), Valid: true},
		},
		ApplicationID: result.ID,
		Application:   result,
	}
	savetx = db.Omit(clause.Associations).Create(&version)
	if savetx.Error != nil {
		return Application{}, savetx.Error
	}

	var adjustment = ApplicationAdjustment{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ReviewableAdjustmentBase: ReviewableAdjustmentBase{
			AdjustmentNumber: 1,
			ReviewState:      reviewstate.Approved,
		},
		ApplicationVersionID: version.ID,
		ApplicationVersion:   version,
		DisplayName:          "App 1",
	}
	if adjustmentCustomizeFunc != nil {
		adjustmentCustomizeFunc(&adjustment)
	}
	savetx = db.Omit(clause.Associations).Create(&adjustment)
	if savetx.Error != nil {
		return Application{}, savetx.Error
	}

	result.LatestVersion = &version
	result.LatestAdjustment = &adjustment
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
func CreateMockRulesetWith1Version(db *gorm.DB, organization Organization, id string, customizeFunc func(*ApprovalRulesetAdjustment)) (ApprovalRuleset, error) {
	var savetx *gorm.DB
	var versionNumber uint32 = 1

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

	version := ApprovalRulesetVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ReviewableVersionBase: ReviewableVersionBase{
			VersionNumber: &versionNumber,
			ApprovedAt:    sql.NullTime{Time: time.Now(), Valid: true},
		},
		ApprovalRulesetID: ruleset.ID,
		ApprovalRuleset:   ruleset,
	}
	savetx = db.Omit(clause.Associations).Create(&version)
	if savetx.Error != nil {
		return ApprovalRuleset{}, fmt.Errorf("Error creating ApprovalRulesetVersion: %w", savetx.Error)
	}

	adjustment := ApprovalRulesetAdjustment{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ReviewableAdjustmentBase: ReviewableAdjustmentBase{
			AdjustmentNumber: 1,
			ReviewState:      reviewstate.Approved,
		},
		ApprovalRulesetVersionID: version.ID,
		ApprovalRulesetVersion:   version,
		DisplayName:              "Ruleset",
		Description:              "",
	}
	if customizeFunc != nil {
		customizeFunc(&adjustment)
	}
	savetx = db.Omit(clause.Associations).Create(&adjustment)
	if savetx.Error != nil {
		return ApprovalRuleset{}, fmt.Errorf("Error creating ApprovalRulesetAdjustment: %w", savetx.Error)
	}

	ruleset.LatestVersion = &version
	ruleset.LatestAdjustment = &adjustment
	return ruleset, nil
}

// CreateMockApplicationRulesetBindingWithEnforcingMode1Version ...
func CreateMockApplicationRulesetBindingWithEnforcingMode1Version(db *gorm.DB, organization Organization, application Application, ruleset ApprovalRuleset, customizeFunc func(*ApplicationApprovalRulesetBindingAdjustment)) (ApplicationApprovalRulesetBinding, error) {
	var binding ApplicationApprovalRulesetBinding
	var version ApplicationApprovalRulesetBindingVersion
	var adjustment ApplicationApprovalRulesetBindingAdjustment
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
		Application:      application,
		ApprovalRuleset:  ruleset,
		LatestVersion:    &version,
		LatestAdjustment: &adjustment,
	}
	savetx := db.Omit(clause.Associations).Create(&binding)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBinding: %w", savetx.Error)
	}

	var versionNumber uint32 = 1
	version, err = CreateMockApplicationApprovalRulesetBindingVersion(db, organization, application, binding, &versionNumber)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBindingVersion: %w", err)
	}

	adjustment, err = CreateMockApplicationApprovalRulesetBindingAdjustment(db, organization, version, customizeFunc)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBindingAdjustment: %w", savetx.Error)
	}

	return binding, nil
}

// CreateMockApplicationApprovalRulesetBindingVersion ...
func CreateMockApplicationApprovalRulesetBindingVersion(db *gorm.DB, organization Organization, application Application, binding ApplicationApprovalRulesetBinding, versionNumber *uint32) (ApplicationApprovalRulesetBindingVersion, error) {
	result := ApplicationApprovalRulesetBindingVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID:     application.ID,
		ApprovalRulesetID: binding.ApprovalRulesetID,
		ReviewableVersionBase: ReviewableVersionBase{
			VersionNumber: versionNumber,
			ApprovedAt:    sql.NullTime{Time: time.Now(), Valid: versionNumber != nil},
		},
		ApplicationApprovalRulesetBinding: binding,
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBindingVersion{}, savetx.Error
	}

	return result, nil
}

func CreateMockApplicationApprovalRulesetBindingAdjustment(db *gorm.DB, organization Organization, version ApplicationApprovalRulesetBindingVersion,
	customizeFunc func(*ApplicationApprovalRulesetBindingAdjustment)) (ApplicationApprovalRulesetBindingAdjustment, error) {

	result := ApplicationApprovalRulesetBindingAdjustment{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationApprovalRulesetBindingVersionID: version.ID,
		ApplicationApprovalRulesetBindingVersion:   version,
		ReviewableAdjustmentBase: ReviewableAdjustmentBase{
			AdjustmentNumber: 1,
			ReviewState:      reviewstate.Approved,
		},
		Enabled: true,
		Mode:    approvalrulesetbindingmode.Enforcing,
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBindingAdjustment{}, savetx.Error
	}

	return result, nil
}

// CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version creates two rulesets and bindings.
// One binding is in permissive mode, other in enforcing mode.
// Each binding, and each ruleset, has 1 version and is approved.
func CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(db *gorm.DB, organization Organization,
	application Application) (ApplicationApprovalRulesetBinding, ApplicationApprovalRulesetBinding, error) {

	// Ruleset 1: permissive

	ruleset1, err := CreateMockRulesetWith1Version(db, organization, "ruleset1", func(adjustment *ApprovalRulesetAdjustment) {
		adjustment.DisplayName = "Ruleset 1"
	})
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, ApplicationApprovalRulesetBinding{}, err
	}
	ruleset1Binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(db, organization, application,
		ruleset1, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
			adjustment.Mode = approvalrulesetbindingmode.Permissive
		})
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, ApplicationApprovalRulesetBinding{}, err
	}

	// Ruleset 2: enforcing

	ruleset2, err := CreateMockRulesetWith1Version(db, organization, "ruleset2", func(adjustment *ApprovalRulesetAdjustment) {
		adjustment.DisplayName = "Ruleset 2"
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

// CreateMockReleaseRulesetBindingWithEnforcingMode1Version ...
func CreateMockReleaseRulesetBindingWithEnforcingMode1Version(db *gorm.DB, organization Organization, release Release,
	ruleset ApprovalRuleset, rulesetVersion ApprovalRulesetVersion, rulesetAdjustment ApprovalRulesetAdjustment,
	customizeFunc func(*ReleaseApprovalRulesetBinding)) (ReleaseApprovalRulesetBinding, error) {

	result := ReleaseApprovalRulesetBinding{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID:                   release.ApplicationID,
		ReleaseID:                       release.ID,
		Release:                         release,
		ApprovalRulesetID:               ruleset.ID,
		ApprovalRuleset:                 ruleset,
		ApprovalRulesetVersionID:        rulesetVersion.ID,
		ApprovalRulesetVersion:          rulesetVersion,
		ApprovalRulesetAdjustmentNumber: rulesetAdjustment.AdjustmentNumber,
		ApprovalRulesetAdjustment:       rulesetAdjustment,
		Mode:                            approvalrulesetbindingmode.Enforcing,
	}
	savetx := db.Omit(clause.Associations).Create(&result)
	if savetx.Error != nil {
		return ReleaseApprovalRulesetBinding{}, savetx.Error
	}
	return result, nil
}

// CreateMockReleaseBackgroundJob ...
func CreateMockReleaseBackgroundJob(db *gorm.DB, organization Organization, app Application, release Release, customizeFunc func(job *ReleaseBackgroundJob)) (ReleaseBackgroundJob, error) {
	result := ReleaseBackgroundJob{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID: app.ID,
		LockSubID:     1,
		ReleaseID:     release.ID,
		Release:       release,
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	tx := db.Omit(clause.Associations).Create(&result)
	if tx.Error != nil {
		return ReleaseBackgroundJob{}, tx.Error
	}
	return result, nil
}

// CreateMockScheduleApprovalRuleWholeDay ...
func CreateMockScheduleApprovalRuleWholeDay(db *gorm.DB, organization Organization, rulesetVersionID uint64,
	rulesetAdjustment ApprovalRulesetAdjustment, customizeFunc func(*ScheduleApprovalRule)) (ScheduleApprovalRule, error) {

	result := ScheduleApprovalRule{
		ApprovalRule: ApprovalRule{
			BaseModel: BaseModel{
				OrganizationID: organization.ID,
				Organization:   organization,
			},
			ApprovalRulesetVersionID:        rulesetVersionID,
			ApprovalRulesetAdjustmentNumber: rulesetAdjustment.AdjustmentNumber,
			ApprovalRulesetAdjustment:       rulesetAdjustment,
			Enabled:                         true,
		},
		BeginTime: sql.NullString{String: "0:00:00", Valid: true},
		EndTime:   sql.NullString{String: "23:59:59", Valid: true},
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	tx := db.Omit(clause.Associations).Create(&result)
	if tx.Error != nil {
		return ScheduleApprovalRule{}, tx.Error
	}
	return result, nil
}

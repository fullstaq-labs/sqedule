package dbmodels

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
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

func CreateMockApplication(db *gorm.DB, organization Organization, customizeFunc func(app *Application)) (Application, error) {
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
	return result, nil
}

func CreateMockApplicationWith1Version(db *gorm.DB, organization Organization, customizeFunc func(app *Application), adjustmentCustomizeFunc func(adjustment *ApplicationAdjustment)) (Application, error) {
	app, err := CreateMockApplication(db, organization, customizeFunc)
	if err != nil {
		return Application{}, err
	}

	version, err := CreateMockApplicationVersion(db, app, lib.NewUint32Ptr(1), nil)
	if err != nil {
		return Application{}, err
	}

	adjustment, err := CreateMockApplicationAdjustment(db, version, 1, adjustmentCustomizeFunc)
	if err != nil {
		return Application{}, err
	}

	app.Version = &version
	app.Version.Adjustment = &adjustment
	return app, nil
}

func CreateMockApplicationVersion(db *gorm.DB, app Application, number *uint32, customizeFunc func(version *ApplicationVersion)) (ApplicationVersion, error) {
	version := ApplicationVersion{
		BaseModel:     app.BaseModel,
		ApplicationID: app.ID,
		Application:   app,
	}
	if number != nil {
		version.VersionNumber = number
		version.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}
	if customizeFunc != nil {
		customizeFunc(&version)
	}
	tx := db.Omit(clause.Associations).Create(&version)
	if tx.Error != nil {
		return ApplicationVersion{}, tx.Error
	}
	return version, nil
}

func CreateMockApplicationAdjustment(db *gorm.DB, version ApplicationVersion, number uint32, customizeFunc func(adjustment *ApplicationAdjustment)) (ApplicationAdjustment, error) {
	adjustment := ApplicationAdjustment{
		BaseModel: version.BaseModel,
		ReviewableAdjustmentBase: ReviewableAdjustmentBase{
			AdjustmentNumber: number,
			ReviewState:      reviewstate.Approved,
		},
		ApplicationVersionID: version.ID,
		ApplicationVersion:   version,
		DisplayName:          "App 1",
	}
	if customizeFunc != nil {
		customizeFunc(&adjustment)
	}
	tx := db.Omit(clause.Associations).Create(&adjustment)
	if tx.Error != nil {
		return ApplicationAdjustment{}, tx.Error
	}
	return adjustment, nil
}

// CreateMockReleaseWithInProgressState ...
func CreateMockReleaseWithInProgressState(db *gorm.DB, organization Organization, application Application,
	customizeFunc func(release *Release)) (Release, error) {

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

func CreateMockApprovalRulesetWith1Version(db *gorm.DB, organization Organization, id string, customizeFunc func(adjustment *ApprovalRulesetAdjustment)) (ApprovalRuleset, error) {
	ruleset, err := CreateMockApprovalRuleset(db, organization, id, nil)
	if err != nil {
		return ApprovalRuleset{}, nil
	}

	version, err := CreateMockApprovalRulesetVersion(db, ruleset, lib.NewUint32Ptr(1), nil)
	if err != nil {
		return ApprovalRuleset{}, nil
	}

	adjustment, err := CreateMockApprovalRulesetAdjustment(db, version, 1, customizeFunc)
	if err != nil {
		return ApprovalRuleset{}, nil
	}

	ruleset.Version = &version
	ruleset.Version.Adjustment = &adjustment
	return ruleset, nil
}

func CreateMockApprovalRuleset(db *gorm.DB, organization Organization, id string, customizeFunc func(ruleset *ApprovalRuleset)) (ApprovalRuleset, error) {
	ruleset := ApprovalRuleset{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ID: id,
	}
	tx := db.Omit(clause.Associations).Create(&ruleset)
	if tx.Error != nil {
		return ApprovalRuleset{}, tx.Error
	}
	return ruleset, nil
}

func CreateMockApprovalRulesetVersion(db *gorm.DB, ruleset ApprovalRuleset, number *uint32, customizeFunc func(version *ApprovalRulesetVersion)) (ApprovalRulesetVersion, error) {
	version := ApprovalRulesetVersion{
		BaseModel:         ruleset.BaseModel,
		ApprovalRulesetID: ruleset.ID,
		ApprovalRuleset:   ruleset,
	}
	if number != nil {
		version.VersionNumber = number
		version.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}
	if customizeFunc != nil {
		customizeFunc(&version)
	}
	tx := db.Omit(clause.Associations).Create(&version)
	if tx.Error != nil {
		return ApprovalRulesetVersion{}, tx.Error
	}
	return version, nil
}

func CreateMockApprovalRulesetAdjustment(db *gorm.DB, version ApprovalRulesetVersion, number uint32, customizeFunc func(adjustment *ApprovalRulesetAdjustment)) (ApprovalRulesetAdjustment, error) {
	adjustment := ApprovalRulesetAdjustment{
		BaseModel: version.BaseModel,
		ReviewableAdjustmentBase: ReviewableAdjustmentBase{
			AdjustmentNumber: number,
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
	tx := db.Omit(clause.Associations).Create(&adjustment)
	if tx.Error != nil {
		return ApprovalRulesetAdjustment{}, tx.Error
	}
	return adjustment, nil
}

func CreateMockApplicationRulesetBindingWithEnforcingMode(db *gorm.DB, organization Organization, application Application, ruleset ApprovalRuleset, customizeFunc func(binding *ApplicationApprovalRulesetBinding)) (ApplicationApprovalRulesetBinding, error) {
	binding := ApplicationApprovalRulesetBinding{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationApprovalRulesetBindingPrimaryKey: ApplicationApprovalRulesetBindingPrimaryKey{
			ApplicationID:     application.ID,
			ApprovalRulesetID: ruleset.ID,
		},
		Application:     application,
		ApprovalRuleset: ruleset,
	}
	if customizeFunc != nil {
		customizeFunc(&binding)
	}
	savetx := db.Omit(clause.Associations).Create(&binding)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBinding: %w", savetx.Error)
	}
	return binding, nil
}

func CreateMockApplicationRulesetBindingWithEnforcingMode1Version(db *gorm.DB, organization Organization, application Application, ruleset ApprovalRuleset, customizeFunc func(*ApplicationApprovalRulesetBindingAdjustment)) (ApplicationApprovalRulesetBinding, error) {
	binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode(db, organization, application, ruleset, nil)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, nil
	}

	version, err := CreateMockApplicationApprovalRulesetBindingVersion(db, organization, application, binding, lib.NewUint32Ptr(1), nil)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBindingVersion: %w", err)
	}

	adjustment, err := CreateMockApplicationApprovalRulesetBindingAdjustment(db, organization, version, customizeFunc)
	if err != nil {
		return ApplicationApprovalRulesetBinding{}, fmt.Errorf("Error creating ApplicationApprovalRulesetBindingAdjustment: %w", err)
	}

	binding.Version = &version
	binding.Version.Adjustment = &adjustment

	return binding, nil
}

func CreateMockApplicationApprovalRulesetBindingVersion(db *gorm.DB, organization Organization, application Application, binding ApplicationApprovalRulesetBinding, number *uint32,
	customizeFunc func(version *ApplicationApprovalRulesetBindingVersion)) (ApplicationApprovalRulesetBindingVersion, error) {

	version := ApplicationApprovalRulesetBindingVersion{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
		ApplicationID:     application.ID,
		ApprovalRulesetID: binding.ApprovalRulesetID,
		ReviewableVersionBase: ReviewableVersionBase{
			VersionNumber: number,
			ApprovedAt:    sql.NullTime{Time: time.Now(), Valid: number != nil},
		},
		ApplicationApprovalRulesetBinding: binding,
	}
	if customizeFunc != nil {
		customizeFunc(&version)
	}
	savetx := db.Omit(clause.Associations).Create(&version)
	if savetx.Error != nil {
		return ApplicationApprovalRulesetBindingVersion{}, savetx.Error
	}
	return version, nil
}

func CreateMockApplicationApprovalRulesetBindingAdjustment(db *gorm.DB, organization Organization, version ApplicationApprovalRulesetBindingVersion,
	customizeFunc func(adjustment *ApplicationApprovalRulesetBindingAdjustment)) (ApplicationApprovalRulesetBindingAdjustment, error) {

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
		Enabled: lib.NewBoolPtr(true),
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

	ruleset1, err := CreateMockApprovalRulesetWith1Version(db, organization, "ruleset1", func(adjustment *ApprovalRulesetAdjustment) {
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

	ruleset2, err := CreateMockApprovalRulesetWith1Version(db, organization, "ruleset2", func(adjustment *ApprovalRulesetAdjustment) {
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

func CreateMockReleaseRulesetBindingWithEnforcingMode(db *gorm.DB, organization Organization, release Release,
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
	rulesetAdjustment ApprovalRulesetAdjustment, customizeFunc func(rule *ScheduleApprovalRule)) (ScheduleApprovalRule, error) {

	result := ScheduleApprovalRule{
		ApprovalRule: ApprovalRule{
			BaseModel: BaseModel{
				OrganizationID: organization.ID,
				Organization:   organization,
			},
			ApprovalRulesetVersionID:        rulesetVersionID,
			ApprovalRulesetAdjustmentNumber: rulesetAdjustment.AdjustmentNumber,
			ApprovalRulesetAdjustment:       rulesetAdjustment,
			Enabled:                         lib.NewBoolPtr(true),
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

func CreateMockCreationAuditRecord(db *gorm.DB, organization Organization, customizeFunc func(record *CreationAuditRecord)) (CreationAuditRecord, error) {
	result := CreationAuditRecord{
		BaseModel: BaseModel{
			OrganizationID: organization.ID,
			Organization:   organization,
		},
	}
	if customizeFunc != nil {
		customizeFunc(&result)
	}
	tx := db.Omit(clause.Associations).Create(&result)
	if tx.Error != nil {
		return CreationAuditRecord{}, tx.Error
	}
	return result, nil
}

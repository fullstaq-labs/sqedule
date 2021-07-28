package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

type Application struct {
	BaseModel
	ID string `gorm:"type:citext; primaryKey; not null"`
	ReviewableBase

	Version *ApplicationVersion `gorm:"-"`
}

type ApplicationVersion struct {
	BaseModel
	ReviewableVersionBase
	ApplicationID string      `gorm:"type:citext; not null"`
	Application   Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	Adjustment *ApplicationAdjustment `gorm:"-"`
}

type ApplicationAdjustment struct {
	BaseModel
	ApplicationVersionID uint64 `gorm:"primaryKey; not null"`
	ReviewableAdjustmentBase
	Enabled *bool `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`

	ApplicationVersion ApplicationVersion `gorm:"foreignKey:OrganizationID,ApplicationVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

//
// ******** Application methods ********
//

// NewDraftVersion returns an unsaved ApplicationVersion and ApplicationAdjustment
// in draft proposal state. Their contents are identical to the currently loaded Version and Adjustment.
func (app Application) NewDraftVersion() (*ApplicationVersion, *ApplicationAdjustment) {
	var adjustment ApplicationAdjustment
	var version *ApplicationVersion = &adjustment.ApplicationVersion

	if app.Version != nil && app.Version.Adjustment != nil {
		adjustment = *app.Version.Adjustment
	}

	version.BaseModel = app.BaseModel
	version.ReviewableVersionBase = ReviewableVersionBase{}
	version.Application = app
	version.ApplicationID = app.ID
	version.Adjustment = &adjustment

	adjustment.BaseModel = app.BaseModel
	adjustment.ApplicationVersionID = 0
	adjustment.ReviewableAdjustmentBase = ReviewableAdjustmentBase{
		AdjustmentNumber: 1,
		ReviewState:      reviewstate.Draft,
	}

	return version, &adjustment
}

func (app Application) CheckNewProposalsRequireReview(action ReviewableAction) bool {
	return true
}

//
// ******** ApplicationAdjustment methods ********
//

func (adjustment ApplicationAdjustment) IsEnabled() bool {
	return lib.DerefBoolPtrWithDefault(adjustment.Enabled, true)
}

//
// ******** Find/load functions ********
//

func FindApplications(db *gorm.DB, organizationID string) ([]Application, error) {
	var result []Application
	tx := db.Where("organization_id = ?", organizationID)
	tx = tx.Find(&result)
	return result, tx.Error
}

func FindApplicationsWithApprovalRuleset(db *gorm.DB, organizationID string, approvalRulesetID string) ([]Application, error) {
	var result []Application
	tx := db.
		Table("application_approval_ruleset_bindings").
		Select("applications.*").
		Joins("LEFT JOIN applications "+
			"ON applications.id = application_approval_ruleset_bindings.application_id "+
			"AND applications.organization_id = application_approval_ruleset_bindings.organization_id").
		Where("application_approval_ruleset_bindings.organization_id = ? "+
			"AND application_approval_ruleset_bindings.approval_ruleset_id = ?",
			organizationID, approvalRulesetID)
	tx = tx.Find(&result)
	return result, tx.Error
}

// FindApplication looks up an Application by its ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindApplication(db *gorm.DB, organizationID string, id string) (Application, error) {
	var result Application

	tx := db.Where("organization_id = ? AND id = ?", organizationID, id)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func FindApplicationVersionByID(db *gorm.DB, organizationID string, applicationID string, versionID uint64) (ApplicationVersion, error) {
	var result ApplicationVersion

	tx := db.Where("organization_id = ? AND application_id = ? AND id = ?", organizationID, applicationID, versionID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

func LoadApplicationsLatestVersionsAndAdjustments(db *gorm.DB, organizationID string, applications []*Application) error {
	err := LoadApplicationsLatestVersions(db, organizationID, applications)
	if err != nil {
		return err
	}

	return LoadApplicationVersionsLatestAdjustments(db, organizationID, CollectApplicationVersions(applications))
}

func LoadApplicationsLatestVersions(db *gorm.DB, organizationID string, applications []*Application) error {
	reviewables := make([]IReviewable, 0, len(applications))
	for _, app := range applications {
		reviewables = append(reviewables, app)
	}

	return LoadReviewablesLatestVersions(
		db,
		organizationID,
		reviewables,
		reflect.TypeOf(ApplicationVersion{}),
		[]string{"application_id"},
	)
}

func LoadApplicationVersionsLatestAdjustments(db *gorm.DB, organizationID string, versions []*ApplicationVersion) error {
	iversions := make([]IReviewableVersion, 0, len(versions))
	for _, version := range versions {
		iversions = append(iversions, version)
	}

	return LoadReviewableVersionsLatestAdjustments(
		db,
		organizationID,
		iversions,
		reflect.TypeOf(ApplicationAdjustment{}),
		"application_version_id",
	)
}

//
// ******** Other functions ********
//

// MakeApplicationsPointerArray returns a `[]Application` into a `[]*Application`.
func MakeApplicationsPointerArray(apps []Application) []*Application {
	result := make([]*Application, 0, len(apps))
	for i := range apps {
		result = append(result, &apps[i])
	}
	return result
}

func CollectApplicationsWithApplicationApprovalRulesetBindings(bindings []ApplicationApprovalRulesetBinding) []*Application {
	result := make([]*Application, 0, len(bindings))
	for i := range bindings {
		binding := &bindings[i]
		result = append(result, &binding.Application)
	}
	return result
}

func CollectApplicationsWithReleases(releases []*Release) []*Application {
	result := make([]*Application, 0, len(releases))
	for _, release := range releases {
		result = append(result, &release.Application)
	}
	return result
}

func CollectApplicationIDs(applications []Application) []string {
	result := make([]string, 0, len(applications))
	for _, app := range applications {
		result = append(result, app.ID)
	}
	return result
}

// CollectApplicationVersions turns a `[]*Application` into a list of their associated ApplicationVersions.
// It does not include nils.
func CollectApplicationVersions(applications []*Application) []*ApplicationVersion {
	result := make([]*ApplicationVersion, 0, len(applications))
	for _, elem := range applications {
		if elem.Version != nil {
			result = append(result, elem.Version)
		}
	}
	return result
}

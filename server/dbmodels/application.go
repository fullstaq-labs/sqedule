package dbmodels

import (
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

type Application struct {
	BaseModel
	ID string `gorm:"type:citext; primaryKey; not null"`
	ReviewableBase
	LatestVersion    *ApplicationVersion    `gorm:"-"`
	LatestAdjustment *ApplicationAdjustment `gorm:"-"`
}

type ApplicationVersion struct {
	BaseModel
	ReviewableVersionBase
	ApplicationID string      `gorm:"type:citext; not null"`
	Application   Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

type ApplicationAdjustment struct {
	BaseModel
	ApplicationVersionID uint64 `gorm:"primaryKey; not null"`
	ReviewableAdjustmentBase
	Enabled bool `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`

	ApplicationVersion ApplicationVersion `gorm:"foreignKey:OrganizationID,ApplicationVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// FindAllApplications ...
func FindAllApplications(db *gorm.DB, organizationID string) ([]Application, error) {
	var result []Application
	tx := db.Where("organization_id = ?", organizationID)
	tx = tx.Find(&result)
	return result, tx.Error
}

func FindAllApplicationsWithApprovalRuleset(db *gorm.DB, organizationID string, approvalRulesetID string) ([]Application, error) {
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

// MakeApplicationsPointerArray ...
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

func CollectApplicationIDs(apps []Application) []string {
	result := make([]string, 0, len(apps))
	for _, app := range apps {
		result = append(result, app.ID)
	}
	return result
}

// LoadApplicationsLatestVersions ...
func LoadApplicationsLatestVersions(db *gorm.DB, organizationID string, applications []*Application) error {
	reviewables := make([]IReviewable, 0, len(applications))
	for _, app := range applications {
		reviewables = append(reviewables, app)
	}

	return LoadReviewablesLatestVersions(
		db,
		reflect.TypeOf(Application{}.ID),
		[]string{"application_id"},
		reflect.TypeOf(Application{}.ID),
		reflect.TypeOf(ApplicationVersion{}),
		reflect.TypeOf(ApplicationVersion{}.ID),
		"application_version_id",
		reflect.TypeOf(ApplicationAdjustment{}),
		organizationID,
		reviewables,
	)
}

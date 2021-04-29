package dbmodels

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

// Application ...
type Application struct {
	BaseModel
	ID                 string                   `gorm:"type:citext; primaryKey; not null"`
	CreatedAt          time.Time                `gorm:"not null"`
	LatestMajorVersion *ApplicationMajorVersion `gorm:"-"`
	LatestMinorVersion *ApplicationMinorVersion `gorm:"-"`
}

// ApplicationMajorVersion ...
type ApplicationMajorVersion struct {
	OrganizationID string       `gorm:"type:citext; primaryKey; not null; index:application_major_version_idx,unique"`
	Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ID             uint64       `gorm:"primaryKey; autoIncrement; not null"`
	ApplicationID  string       `gorm:"type:citext; not null; index:application_major_version_idx,unique"`
	VersionNumber  *uint32      `gorm:"type:int; index:application_major_version_idx,sort:desc,where:version_number IS NOT NULL,unique; check:(version_number > 0)"`
	CreatedAt      time.Time    `gorm:"not null"`
	UpdatedAt      time.Time    `gorm:"not null"`

	Application Application `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

// ApplicationMinorVersion ...
type ApplicationMinorVersion struct {
	BaseModel
	ApplicationMajorVersionID uint64            `gorm:"primaryKey; not null"`
	VersionNumber             uint32            `gorm:"type:int; primaryKey; not null; check:(version_number > 0)"`
	ReviewState               reviewstate.State `gorm:"type:review_state; not null"`
	ReviewComments            sql.NullString
	CreatedAt                 time.Time `gorm:"not null"`
	Enabled                   bool      `gorm:"not null; default:true"`

	DisplayName string `gorm:"not null"`

	ApplicationMajorVersion ApplicationMajorVersion `gorm:"foreignKey:OrganizationID,ApplicationMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
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
		reflect.TypeOf(ApplicationMajorVersion{}),
		reflect.TypeOf(ApplicationMajorVersion{}.ID),
		"application_major_version_id",
		reflect.TypeOf(ApplicationMinorVersion{}),
		organizationID,
		reviewables,
	)
}

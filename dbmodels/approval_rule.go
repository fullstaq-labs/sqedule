package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/approvalpolicy"
	"github.com/fullstaq-labs/sqedule/dbmodels/retrypolicy"
	"gorm.io/gorm"
)

// ApprovalRuleVersionKey ...
type ApprovalRuleVersionKey struct {
	MajorVersionID     uint64
	MinorVersionNumber uint32
}

// ApprovalRule ...
type ApprovalRule struct {
	BaseModel
	ID                                uint64                      `gorm:"primaryKey; autoIncrement; not null"`
	ApprovalRulesetMajorVersionID     uint64                      `gorm:"not null"`
	ApprovalRulesetMinorVersionNumber uint32                      `gorm:"not null"`
	ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Enabled                           bool                        `gorm:"not null; default:true"`
	CreatedAt                         time.Time                   `gorm:"not null"`
}

// HTTPApiApprovalRule ...
type HTTPApiApprovalRule struct {
	ApprovalRule
	URL              string `gorm:"not null"`
	Username         sql.NullString
	Password         sql.NullString
	TLSCaCertificate sql.NullString
	RetryPolicy      retrypolicy.Policy `gorm:"type:retry_policy; not null"`
	RetryLimit       int                `gorm:"not null; default:1; check:((retry_policy = 'retry_on_fail') = (retry_limit IS NOT NULL))"`
}

// ScheduleApprovalRule ...
type ScheduleApprovalRule struct {
	ApprovalRule
	BeginTime    sql.NullString `gorm:"check:((begin_time IS NULL) = (end_time IS NULL))"`
	EndTime      sql.NullString
	DaysOfWeek   sql.NullString
	DaysOfMonth  sql.NullString
	MonthsOfYear sql.NullString
}

// ManualApprovalRule ...
type ManualApprovalRule struct {
	ApprovalRule
	ApprovalPolicy approvalpolicy.Policy `gorm:"type:approval_policy; not null"`
	Minimum        sql.NullInt32         `gorm:"check:((approval_policy = 'minimum') = (minimum IS NOT NULL))"`
}

// FindAllScheduleApprovalRulesBelongingToVersions ...
func FindAllScheduleApprovalRulesBelongingToVersions(db *gorm.DB, organizationID string, versionKeys []ApprovalRuleVersionKey) ([]ScheduleApprovalRule, error) {
	var result []ScheduleApprovalRule
	var versionKeyConditions *gorm.DB = db

	for _, versionKey := range versionKeys {
		versionKeyConditions = versionKeyConditions.Or(
			db.Where("approval_ruleset_major_version_id = ? AND approval_ruleset_minor_version_number = ?",
				versionKey.MajorVersionID, versionKey.MinorVersionNumber),
		)
	}

	tx := db.Where("organization_id = ?", organizationID).Where(versionKeyConditions)
	tx = tx.Find(&result)
	return result, tx.Error
}

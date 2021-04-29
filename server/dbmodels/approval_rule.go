package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalpolicy"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/retrypolicy"
	"gorm.io/gorm"
)

const NumApprovalRuleTypes = 3

// ApprovalRule ...
type ApprovalRule struct {
	BaseModel
	ID                                uint64                      `gorm:"primaryKey; autoIncrement; not null"`
	ApprovalRulesetMajorVersionID     uint64                      `gorm:"not null"`
	ApprovalRulesetMinorVersionNumber uint32                      `gorm:"type:int; not null; check:(approval_ruleset_minor_version_number >= 0)"`
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

type ApprovalRulesetContents struct {
	HTTPApiApprovalRules  []HTTPApiApprovalRule
	ScheduleApprovalRules []ScheduleApprovalRule
	ManualApprovalRules   []ManualApprovalRule
}

func FindAllApprovalRulesWithRuleset(db *gorm.DB, organizationID string, rulesetVersionKey ApprovalRulesetVersionKey) (ApprovalRulesetContents, error) {
	var result ApprovalRulesetContents
	var query, tx *gorm.DB
	var ruleTypesProcessed uint = 0

	query = db.Where("organization_id = ? AND approval_ruleset_major_version_id = ? AND approval_ruleset_minor_version_number = ?",
		organizationID, rulesetVersionKey.MajorVersionID, rulesetVersionKey.MinorVersionNumber)

	ruleTypesProcessed++
	tx = db.Where(query).Find(&result.HTTPApiApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	ruleTypesProcessed++
	tx = db.Where(query).Find(&result.ScheduleApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	ruleTypesProcessed++
	tx = db.Where(query).Find(&result.ManualApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	if ruleTypesProcessed != NumApprovalRuleTypes {
		panic("Bug: code does not cover all approval rule types")
	}

	return result, nil
}

// FindAllScheduleApprovalRulesBelongingToVersions ...
func FindAllScheduleApprovalRulesBelongingToVersions(db *gorm.DB, conditions *gorm.DB, organizationID string, versionKeys []ApprovalRulesetVersionKey) ([]ScheduleApprovalRule, error) {
	var result []ScheduleApprovalRule
	var versionKeyConditions *gorm.DB = db

	for _, versionKey := range versionKeys {
		versionKeyConditions = versionKeyConditions.Or(
			db.Where("approval_ruleset_major_version_id = ? AND approval_ruleset_minor_version_number = ?",
				versionKey.MajorVersionID, versionKey.MinorVersionNumber),
		)
	}

	tx := db.Where("organization_id = ?", organizationID).Where(versionKeyConditions)
	if conditions != nil {
		tx = tx.Where(conditions)
	}
	tx = tx.Find(&result)
	return result, tx.Error
}

package dbmodels

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalpolicy"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/retrypolicy"
	"gorm.io/gorm"
)

const NumApprovalRuleTypes = 3

type ApprovalRule struct {
	BaseModel
	ID                                uint64                      `gorm:"primaryKey; autoIncrement; not null"`
	ApprovalRulesetMajorVersionID     uint64                      `gorm:"not null"`
	ApprovalRulesetMinorVersionNumber uint32                      `gorm:"type:int; not null; check:(approval_ruleset_minor_version_number >= 0)"`
	ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Enabled                           bool                        `gorm:"not null; default:true"`
	CreatedAt                         time.Time                   `gorm:"not null"`

	// BindingMode is the mode with which the containing Ruleset is bound to some entity.
	// This is only set by `FindApprovalRulesBoundToRelease()`. It's not a real table
	// column, so don't add to database migrations.
	BindingMode approvalrulesetbindingmode.Mode `gorm:"<-:false"`
}

type HTTPApiApprovalRule struct {
	ApprovalRule
	URL              string `gorm:"not null"`
	Username         sql.NullString
	Password         sql.NullString
	TLSCaCertificate sql.NullString
	RetryPolicy      retrypolicy.Policy `gorm:"type:retry_policy; not null"`
	RetryLimit       int                `gorm:"not null; default:1; check:((retry_policy = 'retry_on_fail') = (retry_limit IS NOT NULL))"`
}

type ScheduleApprovalRule struct {
	ApprovalRule
	BeginTime    sql.NullString `gorm:"check:((begin_time IS NULL) = (end_time IS NULL))"`
	EndTime      sql.NullString
	DaysOfWeek   sql.NullString
	DaysOfMonth  sql.NullString
	MonthsOfYear sql.NullString
}

type ManualApprovalRule struct {
	ApprovalRule
	ApprovalPolicy approvalpolicy.Policy `gorm:"type:approval_policy; not null"`
	Minimum        sql.NullInt32         `gorm:"check:((approval_policy = 'minimum') = (minimum IS NOT NULL))"`
}

// FindApprovalRulesBoundToRelease finds all ApprovalRules that are bound to a specific Release.
// It populates the `BindingMode` field so that you know which ApprovalRules are bound
// to the Release through which mode.
func FindApprovalRulesBoundToRelease(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) (ApprovalRulesetContents, error) {
	var result ApprovalRulesetContents
	var tx *gorm.DB
	var ruleTypesProcessed uint = 0

	bindingsCondition := db.Where("organization_id = ? AND application_id = ? AND release_id = ?",
		organizationID, applicationID, releaseID)

	const joinConditionString = "LEFT JOIN %s approval_rules " +
		"ON approval_rules.organization_id = release_approval_ruleset_bindings.organization_id " +
		"AND approval_rules.approval_ruleset_major_version_id = release_approval_ruleset_bindings.approval_ruleset_major_version_id " +
		"AND approval_rules.approval_ruleset_minor_version_number = release_approval_ruleset_bindings.approval_ruleset_minor_version_number"
	const selector = "approval_rules.*, release_approval_ruleset_bindings.mode AS binding_mode"

	ruleTypesProcessed++
	tx = db.Where(bindingsCondition).
		Joins(fmt.Sprintf(joinConditionString, "http_api_approval_rules")).
		Select(selector).
		Find(&result.HTTPApiApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	ruleTypesProcessed++
	tx = db.Where(bindingsCondition).
		Joins(fmt.Sprintf(joinConditionString, "schedule_approval_rules")).
		Select(selector).
		Find(&result.ScheduleApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	ruleTypesProcessed++
	tx = db.Where(bindingsCondition).
		Joins(fmt.Sprintf(joinConditionString, "manual_approval_rules")).
		Select(selector).
		Find(&result.ManualApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	if ruleTypesProcessed != NumApprovalRuleTypes {
		panic("Bug: code does not cover all approval rule types")
	}

	return result, nil
}

func FindApprovalRulesInRulesetVersion(db *gorm.DB, organizationID string, rulesetVersionKey ApprovalRulesetVersionKey) (ApprovalRulesetContents, error) {
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

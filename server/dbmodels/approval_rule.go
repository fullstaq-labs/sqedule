package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalpolicy"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/retrypolicy"
	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

type ApprovalRuleType string

const (
	HTTPApiApprovalRuleType  ApprovalRuleType = "http_api"
	ScheduleApprovalRuleType ApprovalRuleType = "schedule"
	ManualApprovalRuleType   ApprovalRuleType = "manual"

	NumApprovalRuleTypes uint = 3
)

type IApprovalRule interface {
	Type() ApprovalRuleType
	AssociateWithApprovalRulesetAdjustment(adjustment ApprovalRulesetAdjustment)
}

type ApprovalRule struct {
	BaseModel
	ID                              uint64                    `gorm:"primaryKey; autoIncrement; not null"`
	ApprovalRulesetVersionID        uint64                    `gorm:"not null"`
	ApprovalRulesetAdjustmentNumber uint32                    `gorm:"type:int; not null; check:(approval_ruleset_adjustment_number >= 0)"`
	ApprovalRulesetAdjustment       ApprovalRulesetAdjustment `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID,ApprovalRulesetAdjustmentNumber; references:OrganizationID,ApprovalRulesetVersionID,AdjustmentNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Enabled                         *bool                     `gorm:"not null; default:true"`
	CreatedAt                       time.Time                 `gorm:"not null"`

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

//
// ******** ApprovalRule methods ********
//

func (r ApprovalRule) ApprovalRulesetVersionAndAdjustmentKey() ApprovalRulesetVersionAndAdjustmentKey {
	return ApprovalRulesetVersionAndAdjustmentKey{
		VersionID:        r.ApprovalRulesetVersionID,
		AdjustmentNumber: r.ApprovalRulesetAdjustmentNumber,
	}
}

func (r *ApprovalRule) AssociateWithApprovalRulesetAdjustment(adjustment ApprovalRulesetAdjustment) {
	r.ApprovalRulesetVersionID = adjustment.ApprovalRulesetVersionID
	r.ApprovalRulesetAdjustmentNumber = adjustment.AdjustmentNumber
	r.ApprovalRulesetAdjustment = adjustment
}

func (r HTTPApiApprovalRule) Type() ApprovalRuleType {
	return HTTPApiApprovalRuleType
}

func (r ScheduleApprovalRule) Type() ApprovalRuleType {
	return ScheduleApprovalRuleType
}

func (r ManualApprovalRule) Type() ApprovalRuleType {
	return ManualApprovalRuleType
}

//
// ******** Find/load functions ********
//

// FindApprovalRulesBoundToRelease finds all ApprovalRules that are bound to a specific Release.
// It populates the `BindingMode` field so that you know which ApprovalRules are bound
// to the Release through which mode.
func FindApprovalRulesBoundToRelease(db *gorm.DB, organizationID string, applicationID string, releaseID uint64) (ApprovalRulesetContents, error) {
	var result ApprovalRulesetContents
	var tx *gorm.DB
	var ruleTypesProcessed uint = 0

	bindingsCondition := db.Where("approval_rules.organization_id = ? "+
		"AND release_approval_ruleset_bindings.application_id = ? "+
		"AND release_approval_ruleset_bindings.release_id = ?",
		organizationID, applicationID, releaseID)

	const joinConditionString = "LEFT JOIN release_approval_ruleset_bindings " +
		"ON approval_rules.organization_id = release_approval_ruleset_bindings.organization_id " +
		"AND approval_rules.approval_ruleset_version_id = release_approval_ruleset_bindings.approval_ruleset_version_id " +
		"AND approval_rules.approval_ruleset_adjustment_number = release_approval_ruleset_bindings.approval_ruleset_adjustment_number"
	const selector = "approval_rules.*, release_approval_ruleset_bindings.mode AS binding_mode"

	ruleTypesProcessed++
	tx = db.Where(bindingsCondition).
		Joins(joinConditionString).
		Table("http_api_approval_rules approval_rules").
		Select(selector).
		Find(&result.HTTPApiApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	ruleTypesProcessed++
	tx = db.Where(bindingsCondition).
		Joins(joinConditionString).
		Table("schedule_approval_rules approval_rules").
		Select(selector).
		Find(&result.ScheduleApprovalRules)
	if tx.Error != nil {
		return ApprovalRulesetContents{}, tx.Error
	}

	ruleTypesProcessed++
	tx = db.Where(bindingsCondition).
		Joins(joinConditionString).
		Table("manual_approval_rules approval_rules").
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

package dbmodels

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

//
// ******** Types, constants & variables ********
//

type CreationAuditRecord struct {
	BaseModel
	ID                   uint64 `gorm:"primaryKey; not null"`
	OrganizationMemberIP sql.NullString
	CreatedAt            time.Time `gorm:"not null"`

	// Object association

	UserEmail sql.NullString `gorm:"type:citext; check:((CASE WHEN user_email IS NULL THEN 0 ELSE 1 END) + (CASE WHEN service_account_name IS NULL THEN 0 ELSE 1 END) <= 1)"`
	User      User           `gorm:"foreignKey:OrganizationID,UserEmail; references:OrganizationID,Email; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ServiceAccountName sql.NullString `gorm:"type:citext"`
	ServiceAccount     ServiceAccount `gorm:"foreignKey:OrganizationID,ServiceAccountName; references:OrganizationID,Name; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	// Subject association

	ApplicationVersionID        *uint64               `gorm:"check:((CASE WHEN application_adjustment_number IS NULL THEN 0 ELSE 1 END) + (CASE WHEN approval_ruleset_adjustment_number IS NULL THEN 0 ELSE 1 END) + (CASE WHEN manual_approval_rule_outcome_id IS NULL THEN 0 ELSE 1 END) + (CASE WHEN release_created_event_id IS NULL THEN 0 ELSE 1 END) + (CASE WHEN release_cancelled_event_id IS NULL THEN 0 ELSE 1 END) = 1)"`
	ApplicationAdjustmentNumber *uint32               `gorm:"type:int; check:((application_version_id IS NULL) = (application_adjustment_number IS NULL))"`
	ApplicationAdjustment       ApplicationAdjustment `gorm:"foreignKey:OrganizationID,ApplicationVersionID,ApplicationAdjustmentNumber; references:OrganizationID,ApplicationVersionID,AdjustmentNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ApprovalRulesetVersionID        *uint64
	ApprovalRulesetAdjustmentNumber *uint32                   `gorm:"type:int; check:((approval_ruleset_version_id IS NULL) = (approval_ruleset_adjustment_number IS NULL))"`
	ApprovalRulesetAdjustment       ApprovalRulesetAdjustment `gorm:"foreignKey:OrganizationID,ApprovalRulesetVersionID,ApprovalRulesetAdjustmentNumber; references:OrganizationID,ApprovalRulesetVersionID,AdjustmentNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ManualApprovalRuleOutcomeID *uint64
	ManualApprovalRuleOutcome   ManualApprovalRuleOutcome `gorm:"foreignKey:OrganizationID,ManualApprovalRuleOutcomeID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ReleaseCreatedEventID *uint64
	ReleaseCreatedEvent   ReleaseCreatedEvent `gorm:"foreignKey:OrganizationID,ReleaseCreatedEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`

	ReleaseCancelledEventID *uint64
	ReleaseCancelledEvent   ReleaseCancelledEvent `gorm:"foreignKey:OrganizationID,ReleaseCancelledEventID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
}

//
// ******** Constructor functions ********
//

// NewCreationAuditRecord returns an unsaved CreationAuditRecord with the given properties.
func NewCreationAuditRecord(organizationID string, creator IOrganizationMember, creatorIP string) CreationAuditRecord {
	result := CreationAuditRecord{
		BaseModel: BaseModel{OrganizationID: organizationID},
	}
	if creator != nil {
		if user, ok := creator.(User); ok {
			result.UserEmail = sql.NullString{String: user.Email, Valid: true}
			result.User = user
		} else if sa, ok := creator.(ServiceAccount); ok {
			result.ServiceAccountName = sql.NullString{String: sa.Name, Valid: true}
			result.ServiceAccount = sa
		}
	}
	if len(creatorIP) > 0 {
		result.OrganizationMemberIP = sql.NullString{String: creatorIP, Valid: true}
	}
	return result
}

//
// ******** Deletion functions ********
//

func DeleteAuditCreationRecordsForApprovalRulesetProposal(db *gorm.DB, organizationID string, proposalID uint64) error {
	return db.
		Where("organization_id = ? AND approval_ruleset_version_id = ?", organizationID, proposalID).
		Delete(CreationAuditRecord{}).
		Error
}

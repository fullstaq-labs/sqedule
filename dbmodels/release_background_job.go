package dbmodels

import (
	"math"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"gorm.io/gorm"
)

// ReleaseBackgroundJobPostgresLockNamespace is the number at which the PostgreSQL advisory lock should start.
// Given a ReleaseBackgroundJob with a certain LockID, the corresponding advisory lock ID is
// `ReleaseBackgroundJobPostgresLockNamespace + LockID`
const ReleaseBackgroundJobPostgresLockNamespace uint64 = 1 * 0xffffffff

// ReleaseBackgroundJobMaxLockID is the maximum value that `ReleaseBackgroundJob.LockID` may have.
var ReleaseBackgroundJobMaxLockID uint32 = uint32(math.Pow(2, 31)) - 1

// ReleaseBackgroundJob ...
type ReleaseBackgroundJob struct {
	BaseModel
	ApplicationID       string            `gorm:"type:citext; primaryKey; not null"`
	DeploymentRequestID uint64            `gorm:"primaryKey; not null"`
	DeploymentRequest   DeploymentRequest `gorm:"foreignKey:OrganizationID,ApplicationID,DeploymentRequestID; references:OrganizationID,ApplicationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	LockID              uint32            `gorm:"type:int; autoIncrement; unique; not null; check:(lock_id > 0)"`
	CreatedAt           time.Time         `gorm:"not null"`
}

// ReleaseBackgroundJobApprovalRulesetBinding ...
type ReleaseBackgroundJobApprovalRulesetBinding struct {
	BaseModel

	ApplicationID        string               `gorm:"type:citext; primaryKey; not null"`
	DeploymentRequestID  uint64               `gorm:"primaryKey; not null"`
	ReleaseBackgroundJob ReleaseBackgroundJob `gorm:"foreignKey:OrganizationID,ApplicationID,DeploymentRequestID; references:OrganizationID,ApplicationID,DeploymentRequestID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetID string          `gorm:"type:citext; primaryKey; not null"`
	ApprovalRuleset   ApprovalRuleset `gorm:"foreignKey:OrganizationID,ApprovalRulesetID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetMajorVersionID uint64                      `gorm:"not null"`
	ApprovalRulesetMajorVersion   ApprovalRulesetMajorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	ApprovalRulesetMinorVersionNumber uint32                      `gorm:"type:int; not null"`
	ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`

	Mode approvalrulesetbindingmode.Mode `gorm:"type:approval_ruleset_binding_mode; not null"`
}

// FindAllReleaseBackgroundJobApprovalRulesetBindings ...
func FindAllReleaseBackgroundJobApprovalRulesetBindings(db *gorm.DB, organizationID string, applicationID string, deploymentRequestID uint64) ([]ReleaseBackgroundJobApprovalRulesetBinding, error) {
	var result []ReleaseBackgroundJobApprovalRulesetBinding
	tx := db.Where("organization_id = ? AND application_id = ? AND deployment_request_id = ?",
		organizationID, applicationID, deploymentRequestID)
	tx = tx.Find(&result)
	return result, tx.Error
}

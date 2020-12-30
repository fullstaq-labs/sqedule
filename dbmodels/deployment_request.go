package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"gorm.io/gorm"
)

// DeploymentRequest ...
type DeploymentRequest struct {
	BaseModel
	ApplicationID  string                       `gorm:"type:citext; primaryKey; not null"`
	Application    Application                  `gorm:"foreignKey:OrganizationID,ApplicationID; references:OrganizationID,ID; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	ID             uint64                       `gorm:"primaryKey; not null"`
	State          deploymentrequeststate.State `gorm:"type:deployment_request_state; not null"`
	SourceIdentity sql.NullString
	Comments       sql.NullString
	CreatedAt      time.Time `gorm:"not null"`
	UpdatedAt      time.Time `gorm:"not null"`
	FinalizedAt    sql.NullTime
}

// FindAllDeploymentRequests ...
func FindAllDeploymentRequests(db *gorm.DB, organizationID string, applicationID string) ([]DeploymentRequest, error) {
	var result []DeploymentRequest
	tx := db.Where("organization_id = ? AND application_id = ?", organizationID, applicationID).Find(&result)
	return result, tx.Error
}

// FindDeploymentRequest looks up a DeploymentRequest by its ID and its application ID.
// When not found, returns a `gorm.ErrRecordNotFound` error.
func FindDeploymentRequest(db *gorm.DB, organizationID string, applicationID string, deploymentRequestID string) (DeploymentRequest, error) {
	var result DeploymentRequest

	tx := db.Where("organization_id = ? AND application_id = ? AND id = ?", organizationID, applicationID, deploymentRequestID)
	tx.Take(&result)
	return result, dbutils.CreateFindOperationError(tx)
}

// CollectDeploymentRequestApplications ...
func CollectDeploymentRequestApplications(deploymentRequests []DeploymentRequest) []*Application {
	result := make([]*Application, 0)
	for i := range deploymentRequests {
		deploymentRequest := &deploymentRequests[i]
		result = append(result, &deploymentRequest.Application)
	}
	return result
}

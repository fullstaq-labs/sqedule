package httpapi

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
)

type deploymentRequestJSON struct {
	ApplicationID  *string    `json:"application_id"`
	ID             *uint64    `json:"id"`
	State          *string    `json:"state"`
	SourceIdentity *string    `json:"source_identity"`
	Comments       *string    `json:"comments"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
	FinalizedAt    *time.Time `json:"finalized_at"`
}

func createDeploymentRequestJSONFromDbModel(deploymentRequest dbmodels.DeploymentRequest) deploymentRequestJSON {
	state := string(deploymentRequest.State)
	result := deploymentRequestJSON{
		ApplicationID: &deploymentRequest.ApplicationID,
		ID:            &deploymentRequest.ID,
		State:         &state,
		CreatedAt:     &deploymentRequest.CreatedAt,
		UpdatedAt:     &deploymentRequest.UpdatedAt,
	}
	if deploymentRequest.SourceIdentity.Valid {
		result.SourceIdentity = &deploymentRequest.SourceIdentity.String
	}
	if deploymentRequest.Comments.Valid {
		result.Comments = &deploymentRequest.Comments.String
	}
	if deploymentRequest.FinalizedAt.Valid {
		result.FinalizedAt = &deploymentRequest.FinalizedAt.Time
	}
	return result
}

func patchDeploymentRequestDbModelFromJSON(deploymentRequest *dbmodels.DeploymentRequest, json deploymentRequestJSON) {
	if json.SourceIdentity != nil {
		deploymentRequest.SourceIdentity = sql.NullString{*json.SourceIdentity, true}
	}
	if json.Comments != nil {
		deploymentRequest.Comments = sql.NullString{*json.Comments, true}
	}
}

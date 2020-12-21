package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/gin-gonic/gin"
)

// DeploymentRequestAction ...
type DeploymentRequestAction int

const (
	// ActionReadAllDeploymentRequests ...
	ActionReadAllDeploymentRequests = iota
	// ActionReadDeploymentRequest ...
	ActionReadDeploymentRequest
	// ActionUpdateDeploymentRequest ...
	ActionUpdateDeploymentRequest
	// ActionDeleteDeploymentRequest ...
	ActionDeleteDeploymentRequest
)

// GetDeploymentRequestAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target DeploymentRequest.
func GetDeploymentRequestAuthorizations(orgMember dbmodels.IOrganizationMember,
	target dbmodels.DeploymentRequest) map[DeploymentRequestAction]bool {

	result := make(map[DeploymentRequestAction]bool)
	concreteOrgMember := orgMember.GetOrganizationMember()

	result[ActionReadAllDeploymentRequests] = true

	if concreteOrgMember.BaseModel.OrganizationID != target.BaseModel.OrganizationID {
		return result
	}

	result[ActionReadDeploymentRequest] = true
	result[ActionUpdateDeploymentRequest] = true
	result[ActionDeleteDeploymentRequest] = true

	return result
}

// AuthorizeDeploymentRequestAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target DeploymentRequest.
func AuthorizeDeploymentRequestAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember,
	target dbmodels.DeploymentRequest, action DeploymentRequestAction) bool {

	permittedActions := GetDeploymentRequestAuthorizations(orgMember, target)

	if !permittedActions[action] {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

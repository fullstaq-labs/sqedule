package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/gin-gonic/gin"
)

// ApplicationAction ...
type ApplicationAction int

const (
	// ActionCreateApplication ...
	ActionCreateApplication ApplicationAction = iota
	// ActionReadAllApplications ...
	ActionReadAllApplications
	// ActionReadApplication ...
	ActionReadApplication
	// ActionUpdateApplication ...
	ActionUpdateApplication
	// ActionDeleteApplication ...
	ActionDeleteApplication
	// ActionCreateRelease ...
	ActionCreateRelease
)

// GetApplicationAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Application.
func GetApplicationAuthorizations(orgMember dbmodels.IOrganizationMember,
	target dbmodels.Application) map[ApplicationAction]struct{} {

	result := make(map[ApplicationAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.BaseModel.OrganizationID != target.BaseModel.OrganizationID {
		return result
	}

	result[ActionReadAllApplications] = struct{}{}
	result[ActionCreateApplication] = struct{}{}
	result[ActionReadApplication] = struct{}{}
	result[ActionUpdateApplication] = struct{}{}
	result[ActionDeleteApplication] = struct{}{}
	result[ActionCreateRelease] = struct{}{}

	return result
}

// AuthorizeApplicationAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target Application.
func AuthorizeApplicationAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember,
	target dbmodels.Application, action ApplicationAction) bool {

	permittedActions := GetApplicationAuthorizations(orgMember, target)

	if _, ok := permittedActions[action]; !ok {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

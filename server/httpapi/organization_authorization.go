package httpapi

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/organizationmemberrole"
	"github.com/gin-gonic/gin"
)

// OrganizationAction ...
type OrganizationAction int

const (
	// ActionCreateOrganization ...
	ActionCreateOrganization OrganizationAction = iota
	// ActionReadOrganization ...
	ActionReadOrganization
	// ActionUpdateOrganization ...
	ActionUpdateOrganization
	// ActionDeleteOrganization ...
	ActionDeleteOrganization
)

// GetOrganizationAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Organization.
func GetOrganizationAuthorizations(orgMember dbmodels.IOrganizationMember, targetOrganizationID string) map[OrganizationAction]struct{} {
	result := make(map[OrganizationAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.Role == organizationmemberrole.OrgAdmin {
		result[ActionCreateOrganization] = struct{}{}
		result[ActionReadOrganization] = struct{}{}
		result[ActionUpdateOrganization] = struct{}{}
		result[ActionDeleteOrganization] = struct{}{}
		return result
	}

	if concreteOrgMember.BaseModel.OrganizationID != targetOrganizationID {
		return result
	}

	result[ActionReadOrganization] = struct{}{}
	if concreteOrgMember.Role == organizationmemberrole.Admin {
		result[ActionUpdateOrganization] = struct{}{}
	}
	return result
}

// AuthorizeOrganizationAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target Organization.
func AuthorizeOrganizationAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember, targetOrganizationID string,
	action OrganizationAction) bool {

	permittedActions := GetOrganizationAuthorizations(orgMember, targetOrganizationID)

	if _, ok := permittedActions[action]; !ok {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

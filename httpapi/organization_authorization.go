package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/organizationmemberrole"
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
func GetOrganizationAuthorizations(orgMember dbmodels.IOrganizationMember, targetOrganizationID string) map[OrganizationAction]bool {
	result := make(map[OrganizationAction]bool)
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.Role == organizationmemberrole.OrgAdmin {
		result[ActionCreateOrganization] = true
		result[ActionReadOrganization] = true
		result[ActionUpdateOrganization] = true
		result[ActionDeleteOrganization] = true
		return result
	}

	if concreteOrgMember.BaseModel.OrganizationID != targetOrganizationID {
		return result
	}

	result[ActionReadOrganization] = true
	if concreteOrgMember.Role == organizationmemberrole.Admin {
		result[ActionUpdateOrganization] = true
	}
	return result
}

// AuthorizeOrganizationAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target Organization.
func AuthorizeOrganizationAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember, targetOrganizationID string,
	action OrganizationAction) bool {

	permittedActions := GetOrganizationAuthorizations(orgMember, targetOrganizationID)

	if !permittedActions[action] {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

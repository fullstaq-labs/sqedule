package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/organizationmemberrole"
	"github.com/gin-gonic/gin"
)

// OrganizationAction ...
type OrganizationAction int

const (
	// CreateOrganization ...
	CreateOrganization OrganizationAction = iota
	// ReadOrganization ...
	ReadOrganization
	// UpdateOrganization ...
	UpdateOrganization
	// DeleteOrganization ...
	DeleteOrganization
)

// GetOrganizationAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Organization.
func GetOrganizationAuthorizations(orgMember dbmodels.IOrganizationMember, targetOrganizationID string) map[OrganizationAction]bool {
	result := make(map[OrganizationAction]bool)
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.Role == organizationmemberrole.OrgAdmin {
		result[CreateOrganization] = true
		result[ReadOrganization] = true
		result[UpdateOrganization] = true
		result[DeleteOrganization] = true
		return result
	}

	if concreteOrgMember.BaseModel.OrganizationID != targetOrganizationID {
		return result
	}

	result[ReadOrganization] = true
	if concreteOrgMember.Role == organizationmemberrole.Admin {
		result[UpdateOrganization] = true
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

func respondWithUnauthorizedError(ginctx *gin.Context) {
	ginctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized action"})
}

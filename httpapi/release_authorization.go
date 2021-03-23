package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/gin-gonic/gin"
)

// ReleaseAction ...
type ReleaseAction int

const (
	// ActionReadAllReleases ...
	ActionReadAllReleases = iota
	// ActionReadRelease ...
	ActionReadRelease
	// ActionUpdateRelease ...
	ActionUpdateRelease
	// ActionDeleteRelease ...
	ActionDeleteRelease
)

// GetReleaseAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Release.
func GetReleaseAuthorizations(orgMember dbmodels.IOrganizationMember,
	target dbmodels.Release) map[ReleaseAction]struct{} {

	result := make(map[ReleaseAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	result[ActionReadAllReleases] = struct{}{}

	if concreteOrgMember.BaseModel.OrganizationID != target.BaseModel.OrganizationID {
		return result
	}

	result[ActionReadRelease] = struct{}{}
	result[ActionUpdateRelease] = struct{}{}
	result[ActionDeleteRelease] = struct{}{}

	return result
}

// AuthorizeReleaseAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target Release.
func AuthorizeReleaseAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember,
	target dbmodels.Release, action ReleaseAction) bool {

	permittedActions := GetReleaseAuthorizations(orgMember, target)

	if _, ok := permittedActions[action]; !ok {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

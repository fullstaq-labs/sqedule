package authz

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

const (
	ActionListReleases CollectionAction = "releases/list"

	ActionReadRelease   SingularAction = "release/read"
	ActionUpdateRelease SingularAction = "release/update"
	ActionDeleteRelease SingularAction = "release/delete"
)

type ReleaseAuthorizer struct{}

// CollectionAuthorizations returns which collection actions an OrganizationMember is
// allowed to perform.
func (ReleaseAuthorizer) CollectionAuthorizations(orgMember dbmodels.IOrganizationMember) map[CollectionAction]struct{} {
	result := make(map[CollectionAction]struct{})

	result[ActionListReleases] = struct{}{}

	return result
}

// SingularAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Release.
func (ReleaseAuthorizer) SingularAuthorizations(orgMember dbmodels.IOrganizationMember,
	target interface{}) map[SingularAction]struct{} {

	result := make(map[SingularAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.BaseModel.OrganizationID != target.(dbmodels.Release).BaseModel.OrganizationID {
		return result
	}

	result[ActionReadRelease] = struct{}{}
	result[ActionUpdateRelease] = struct{}{}
	result[ActionDeleteRelease] = struct{}{}

	return result
}

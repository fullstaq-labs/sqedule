package authz

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

const (
	ActionCreateApplication CollectionAction = "applications/create"
	ActionListApplications  CollectionAction = "applications/list"

	ActionReadApplication   SingularAction = "application/read"
	ActionUpdateApplication SingularAction = "application/update"
	ActionDeleteApplication SingularAction = "application/delete"

	ActionCreateRelease SingularAction = "release.create"
)

type ApplicationAuthorizer struct{}

// CollectionAuthorizations returns which collection actions an OrganizationMember is
// allowed to perform.
func (ApplicationAuthorizer) CollectionAuthorizations(orgMember dbmodels.IOrganizationMember) map[CollectionAction]struct{} {
	result := make(map[CollectionAction]struct{})

	result[ActionCreateApplication] = struct{}{}
	result[ActionListApplications] = struct{}{}

	return result
}

// SingularAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Application.
func (ApplicationAuthorizer) SingularAuthorizations(orgMember dbmodels.IOrganizationMember,
	target interface{}) map[SingularAction]struct{} {

	result := make(map[SingularAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.BaseModel.OrganizationID != target.(dbmodels.Application).BaseModel.OrganizationID {
		return result
	}

	result[ActionReadApplication] = struct{}{}
	result[ActionUpdateApplication] = struct{}{}
	result[ActionDeleteApplication] = struct{}{}
	result[ActionCreateRelease] = struct{}{}

	return result
}

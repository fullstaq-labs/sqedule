package authz

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/organizationmemberrole"
)

const (
	ActionCreateOrganization CollectionAction = "organization/create"

	ActionReadOrganization   SingularAction = "organization/read"
	ActionUpdateOrganization SingularAction = "organization/update"
	ActionDeleteOrganization SingularAction = "organization/delete"
)

type OrganizationAuthorizer struct{}

// CollectionAuthorizations returns which collection actions an OrganizationMember is
// allowed to perform.
func (OrganizationAuthorizer) CollectionAuthorizations(orgMember dbmodels.IOrganizationMember) map[CollectionAction]struct{} {
	result := make(map[CollectionAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.Role == organizationmemberrole.OrgAdmin {
		result[ActionCreateOrganization] = struct{}{}
	}

	return result
}

// SingularAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target Organization ID.
func (OrganizationAuthorizer) SingularAuthorizations(orgMember dbmodels.IOrganizationMember, targetOrganizationID interface{}) map[SingularAction]struct{} {
	result := make(map[SingularAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.Role == organizationmemberrole.OrgAdmin {
		result[ActionReadOrganization] = struct{}{}
		result[ActionUpdateOrganization] = struct{}{}
		result[ActionDeleteOrganization] = struct{}{}
		return result
	}

	if concreteOrgMember.BaseModel.OrganizationID != targetOrganizationID.(string) {
		return result
	}

	result[ActionReadOrganization] = struct{}{}
	if concreteOrgMember.Role == organizationmemberrole.Admin {
		result[ActionUpdateOrganization] = struct{}{}
	}

	return result
}

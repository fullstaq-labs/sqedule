package authz

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

const (
	ActionListApplicationApprovalRulesetBindings CollectionAction = "application_approval_ruleset_bindings/list"
)

type ApplicationApprovalRulesetBindingAuthorizer struct{}

// CollectionAuthorizations returns which collection actions an OrganizationMember is
// allowed to perform.
func (ApplicationApprovalRulesetBindingAuthorizer) CollectionAuthorizations(orgMember dbmodels.IOrganizationMember) map[CollectionAction]struct{} {
	result := make(map[CollectionAction]struct{})

	result[ActionListApplicationApprovalRulesetBindings] = struct{}{}

	return result
}

// SingularAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target ApplicationApprovalRUlesetBinding.
func (ApplicationApprovalRulesetBindingAuthorizer) SingularAuthorizations(orgMember dbmodels.IOrganizationMember,
	target interface{}) map[SingularAction]struct{} {

	result := make(map[SingularAction]struct{})

	if orgMember.GetOrganizationID() != target.(dbmodels.Application).OrganizationID {
		return result
	}

	return result
}

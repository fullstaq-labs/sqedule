package authz

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
)

const (
	ActionCreateApprovalRuleset CollectionAction = "approval_rulesets/create"
	ActionListApprovalRulesets  CollectionAction = "approval_rulesets/list"

	ActionReadApprovalRuleset   SingularAction = "approval_ruleset/read"
	ActionUpdateApprovalRuleset SingularAction = "approval_ruleset/update"
	ActionDeleteApprovalRuleset SingularAction = "approval_ruleset/delete"
)

type ApprovalRulesetAuthorizer struct{}

// CollectionAuthorizations returns which collection actions an OrganizationMember is
// allowed to perform.
func (ApprovalRulesetAuthorizer) CollectionAuthorizations(orgMember dbmodels.IOrganizationMember) map[CollectionAction]struct{} {
	result := make(map[CollectionAction]struct{})

	result[ActionCreateApprovalRuleset] = struct{}{}
	result[ActionListApprovalRulesets] = struct{}{}

	return result
}

// SingularAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target ApprovalRuleset.
func (ApprovalRulesetAuthorizer) SingularAuthorizations(orgMember dbmodels.IOrganizationMember,
	target interface{}) map[SingularAction]struct{} {

	result := make(map[SingularAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.BaseModel.OrganizationID != target.(dbmodels.ApprovalRuleset).BaseModel.OrganizationID {
		return result
	}

	result[ActionReadApprovalRuleset] = struct{}{}
	result[ActionUpdateApprovalRuleset] = struct{}{}
	result[ActionDeleteApprovalRuleset] = struct{}{}

	return result
}
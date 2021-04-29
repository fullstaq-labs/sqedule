package httpapi

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/gin-gonic/gin"
)

// ApprovalRulesetAction ...
type ApprovalRulesetAction int

const (
	// ActionCreateApplication ...
	ActionCreateApprovalRuleset ApprovalRulesetAction = iota
	// ActionReadAllApprovalRulesets ...
	ActionReadAllApprovalRulesets
	// ActionReadApprovalRuleset ...
	ActionReadApprovalRuleset
	// ActionUpdateApprovalRuleset ...
	ActionUpdateApprovalRuleset
	// ActionDeleteApprovalRuleset ...
	ActionDeleteApprovalRuleset
)

// GetApprovalRulesetAuthorizations returns which actions an OrganizationMember is
// allowed to perform, on a target ApprovalRuleset.
func GetApprovalRulesetAuthorizations(orgMember dbmodels.IOrganizationMember,
	target dbmodels.ApprovalRuleset) map[ApprovalRulesetAction]struct{} {

	result := make(map[ApprovalRulesetAction]struct{})
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.BaseModel.OrganizationID != target.BaseModel.OrganizationID {
		return result
	}

	result[ActionReadAllApprovalRulesets] = struct{}{}
	result[ActionCreateApprovalRuleset] = struct{}{}
	result[ActionReadApprovalRuleset] = struct{}{}
	result[ActionUpdateApprovalRuleset] = struct{}{}
	result[ActionDeleteApprovalRuleset] = struct{}{}

	return result
}

// AuthorizeApprovalRulesetAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target ApprovalRuleset.
func AuthorizeApprovalRulesetAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember,
	target dbmodels.ApprovalRuleset, action ApprovalRulesetAction) bool {

	permittedActions := GetApprovalRulesetAuthorizations(orgMember, target)

	if _, ok := permittedActions[action]; !ok {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

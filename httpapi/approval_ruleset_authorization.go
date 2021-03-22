package httpapi

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
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
	target dbmodels.ApprovalRuleset) map[ApprovalRulesetAction]bool {

	result := make(map[ApprovalRulesetAction]bool)
	concreteOrgMember := orgMember.GetOrganizationMember()

	if concreteOrgMember.BaseModel.OrganizationID != target.BaseModel.OrganizationID {
		return result
	}

	result[ActionReadAllApprovalRulesets] = true
	result[ActionCreateApprovalRuleset] = true
	result[ActionReadApprovalRuleset] = true
	result[ActionUpdateApprovalRuleset] = true
	result[ActionDeleteApprovalRuleset] = true

	return result
}

// AuthorizeApprovalRulesetAction checks whether an OrganizationMember is allowed to
// perform the given action, on a target ApprovalRuleset.
func AuthorizeApprovalRulesetAction(ginctx *gin.Context, orgMember dbmodels.IOrganizationMember,
	target dbmodels.ApprovalRuleset, action ApprovalRulesetAction) bool {

	permittedActions := GetApprovalRulesetAuthorizations(orgMember, target)

	if !permittedActions[action] {
		respondWithUnauthorizedError(ginctx)
		return false
	}

	return true
}

package controllers

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/gin-gonic/gin"
)

func (ctx Context) ListApplicationApprovalRulesetBindings(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bindings, err := dbmodels.FindAllApplicationApprovalRulesetBindings(
		tx.Preload("Application").Preload("ApprovalRuleset"),
		orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(bindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings))
	if err != nil {
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApplicationApprovalRulesetBinding, 0, len(bindings))
	for _, binding := range bindings {
		outputList = append(outputList, json.CreateFromDbApplicationApprovalRulesetBinding(binding,
			*binding.Version, *binding.Version.Adjustment))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

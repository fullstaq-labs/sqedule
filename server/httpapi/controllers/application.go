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

func (ctx Context) GetApplications(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionListApplications) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	apps, err := dbmodels.FindAllApplications(
		tx.Order("created_at DESC"),
		orgID)
	if err != nil {
		respondWithDbQueryError("applications", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, dbmodels.MakeApplicationsPointerArray(apps))
	if err != nil {
		respondWithDbQueryError("application versions", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.Application, 0, len(apps))
	for _, app := range apps {
		outputList = append(outputList, json.CreateFromDbApplication(app, *app.Version, *app.Version.Adjustment, nil))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApplication(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.Application{&app})
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, app) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	bindings, err := dbmodels.FindAllApplicationApprovalRulesetBindings(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(bindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings))
	if err != nil {
		respondWithDbQueryError("approval ruleset versions", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateFromDbApplication(app, *app.Version, *app.Version.Adjustment, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

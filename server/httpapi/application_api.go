package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/gin-gonic/gin"
)

// GetAllApplications ...
func (ctx Context) GetAllApplications(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionListApplications) {
		respondWithUnauthorizedError(ginctx)
		return
	}

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

	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID, dbmodels.MakeApplicationsPointerArray(apps))
	if err != nil {
		respondWithDbQueryError("application versions", err, ginctx)
		return
	}

	outputList := make([]json.Application, 0, len(apps))
	for _, app := range apps {
		outputList = append(outputList, json.CreateFromDbApplication(app, *app.LatestMajorVersion, *app.LatestMinorVersion, nil))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// GetApplication ...
func (ctx Context) GetApplication(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	id := ginctx.Param("application_id")

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID, []*dbmodels.Application{&app})
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, app) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	bindings, err := dbmodels.FindAllApplicationApprovalRulesetBindings(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(bindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings))
	if err != nil {
		respondWithDbQueryError("approval ruleset versions", err, ginctx)
		return
	}

	output := json.CreateFromDbApplication(app, *app.LatestMajorVersion, *app.LatestMinorVersion, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

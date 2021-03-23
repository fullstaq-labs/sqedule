package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
)

// GetAllApplications ...
func (ctx Context) GetAllApplications(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	authTarget := dbmodels.Application{BaseModel: dbmodels.BaseModel{OrganizationID: orgID}}
	if !AuthorizeApplicationAction(ginctx, orgMember, authTarget, ActionReadAllApplications) {
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

	outputList := make([]applicationJSON, 0, len(apps))
	for _, app := range apps {
		outputList = append(outputList, createApplicationJSONFromDbModel(app, *app.LatestMajorVersion, *app.LatestMinorVersion, nil))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// GetApplication ...
func (ctx Context) GetApplication(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
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

	if !AuthorizeApplicationAction(ginctx, orgMember, app, ActionReadApplication) {
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

	output := createApplicationJSONFromDbModel(app, *app.LatestMajorVersion, *app.LatestMinorVersion, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

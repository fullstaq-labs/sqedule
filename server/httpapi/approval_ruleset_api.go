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

// GetAllApprovalRulesets ...
func (ctx Context) GetAllApprovalRulesets(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionListApprovalRulesets) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	pagination, err := dbutils.ParsePaginationOptions(ginctx)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rulesets, err := dbmodels.FindAllApprovalRulesetsWithStats(ctx.Db, orgID, pagination)
	if err != nil {
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithoutStats(rulesets))
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	outputList := make([]json.ApprovalRulesetWithStats, 0, len(rulesets))
	for _, ruleset := range rulesets {
		outputList = append(outputList, json.CreateFromDbApprovalRulesetWithStats(ruleset,
			*ruleset.LatestMajorVersion, *ruleset.LatestMinorVersion))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// GetApprovalRuleset ...
func (ctx Context) GetApprovalRuleset(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	id := ginctx.Param("id")

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		[]*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	rules, err := dbmodels.FindAllApprovalRulesWithRuleset(ctx.Db, orgID, dbmodels.ApprovalRulesetVersionKey{
		MajorVersionID:     ruleset.LatestMajorVersion.ID,
		MinorVersionNumber: ruleset.LatestMinorVersion.VersionNumber,
	})
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	appBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(appBindings))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	releaseBindings, err := dbmodels.FindAllReleaseApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Release.Application"), orgID, ruleset.ID)
	if err != nil {
		respondWithDbQueryError("release approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithReleases(dbmodels.CollectReleasesWithReleaseApprovalRulesetBindings(releaseBindings)))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	output := json.CreateFromDbApprovalRulesetWithBindingAndRuleAssociations(ruleset, *ruleset.LatestMajorVersion,
		*ruleset.LatestMinorVersion, appBindings, releaseBindings, rules)
	ginctx.JSON(http.StatusOK, output)
}

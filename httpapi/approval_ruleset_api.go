package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
)

// GetAllApprovalRulesets ...
func (ctx Context) GetAllApprovalRulesets(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	authTarget := dbmodels.ApprovalRuleset{BaseModel: dbmodels.BaseModel{OrganizationID: orgID}}
	if !AuthorizeApprovalRulesetAction(ginctx, orgMember, authTarget, ActionReadAllApprovalRulesets) {
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

	outputList := make([]approvalRulesetWithStatsJSON, 0, len(rulesets))
	for _, ruleset := range rulesets {
		outputList = append(outputList, createApprovalRulesetWithStatsJSONFromDbModel(ruleset,
			*ruleset.LatestMajorVersion, *ruleset.LatestMinorVersion))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// GetApprovalRuleset ...
func (ctx Context) GetApprovalRuleset(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	id := ginctx.Param("id")

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	if !AuthorizeApprovalRulesetAction(ginctx, orgMember, ruleset, ActionReadApprovalRuleset) {
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

	output := createApprovalRulesetWithBindingAndRuleAssociationsJSONFromDbModel(ruleset, *ruleset.LatestMajorVersion,
		*ruleset.LatestMinorVersion, appBindings, releaseBindings, rules)
	ginctx.JSON(http.StatusOK, output)
}

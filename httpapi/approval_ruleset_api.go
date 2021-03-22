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

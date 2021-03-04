package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
)

// GetAllApplicationApprovalRulesetBindings ...
func (ctx Context) GetAllApplicationApprovalRulesetBindings(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	if !AuthorizeApplicationAction(ginctx, orgMember, application, ActionReadApplication) {
		return
	}

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

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingPointerArray(bindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApplicationApprovalRulesetBindingRulesets(bindings))
	if err != nil {
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
	}

	outputList := make([]applicationApprovalRulesetBindingJSON, 0, len(bindings))
	for _, binding := range bindings {
		outputList = append(outputList, createApplicationApprovalRulesetBindingJSONFromDbModel(binding, false))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

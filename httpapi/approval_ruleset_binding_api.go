package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
)

// GetAllApprovalRulesetBindings ...
func (ctx Context) GetAllApprovalRulesetBindings(ginctx *gin.Context) {
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
	bindings, err := dbmodels.FindAllApprovalRulesetBindings(
		tx.Preload("Application").Preload("ApprovalRuleset"),
		orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetBindingRulesets(bindings))
	if err != nil {
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
	}

	outputList := make([]approvalRulesetBindingJSON, 0, len(bindings))
	for _, binding := range bindings {
		outputList = append(outputList, createApprovalRulesetBindingJSONFromDbModel(binding, false))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

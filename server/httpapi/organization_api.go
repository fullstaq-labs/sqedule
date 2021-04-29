package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/gin-gonic/gin"
)

// GetCurrentOrganization ...
func (ctx Context) GetCurrentOrganization(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

// PatchCurrentOrganization ...
func (ctx Context) PatchCurrentOrganization(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	var input json.Organization
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	json.PatchDbOrganization(&organization, input)
	if err = ctx.Db.Save(&organization).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

// GetOrganization ...
func (ctx Context) GetOrganization(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := ginctx.Param("id")

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

// PatchOrganization ...
func (ctx Context) PatchOrganization(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := ginctx.Param("id")

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	var input json.Organization
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	json.PatchDbOrganization(&organization, input)
	if err = ctx.Db.Save(&organization).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

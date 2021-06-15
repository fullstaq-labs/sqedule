package controllers

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/gin-gonic/gin"
)

func (ctx Context) GetCurrentOrganization(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	// Check authorization

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) GetOrganization(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := ginctx.Param("id")

	// Check authorization

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateCurrentOrganization(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	var input json.Organization
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	// Modify database

	json.PatchDbOrganization(&organization, input)
	if err = ctx.Db.Save(&organization).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateOrganization(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := ginctx.Param("id")

	var input json.Organization
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.OrganizationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateOrganization, orgID) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	// Modify database

	var organization2 dbmodels.Organization = organization
	json.PatchDbOrganization(&organization2, input)
	if err = ctx.Db.Model(&organization).Updates(organization2).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateFromDbOrganization(organization)
	ginctx.JSON(http.StatusOK, output)
}

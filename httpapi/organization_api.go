package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/gin-gonic/gin"
)

// GetCurrentOrganization ...
func (ctx Context) GetCurrentOrganization(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	if !AuthorizeOrganizationAction(ginctx, orgMember, orgID, ActionReadOrganization) {
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	output := createOrganizationJSONFromDbModel(organization)
	ginctx.JSON(http.StatusOK, output)
}

// PatchCurrentOrganization ...
func (ctx Context) PatchCurrentOrganization(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID

	if !AuthorizeOrganizationAction(ginctx, orgMember, orgID, ActionUpdateOrganization) {
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	var input organizationJSON
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	patchOrganizationDbModelFromJSON(&organization, input)
	if err = ctx.Db.Save(&organization).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createOrganizationJSONFromDbModel(organization)
	ginctx.JSON(http.StatusOK, output)
}

// GetOrganization ...
func (ctx Context) GetOrganization(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := ginctx.Param("id")

	if !AuthorizeOrganizationAction(ginctx, orgMember, orgID, ActionReadOrganization) {
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	output := createOrganizationJSONFromDbModel(organization)
	ginctx.JSON(http.StatusOK, output)
}

// PatchOrganization ...
func (ctx Context) PatchOrganization(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := ginctx.Param("id")

	if !AuthorizeOrganizationAction(ginctx, orgMember, orgID, ActionUpdateOrganization) {
		return
	}

	organization, err := dbmodels.FindOrganizationByID(ctx.Db, orgID)
	if err != nil {
		respondWithDbQueryError("organization", err, ginctx)
		return
	}

	var input organizationJSON
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	patchOrganizationDbModelFromJSON(&organization, input)
	if err = ctx.Db.Save(&organization).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createOrganizationJSONFromDbModel(organization)
	ginctx.JSON(http.StatusOK, output)
}

package httpapi

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
	"github.com/gin-gonic/gin"
)

// GetAllDeploymentRequests ...
func (ctx Context) GetAllDeploymentRequests(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	if !AuthorizeDeploymentRequestAction(ginctx, orgMember, dbmodels.DeploymentRequest{},
		ActionReadAllDeploymentRequests) {

		return
	}

	deploymentRequests, err := dbmodels.FindAllDeploymentRequests(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("deployment requests", err, ginctx)
		return
	}

	outputList := make([]deploymentRequestJSON, 0, len(deploymentRequests))
	for _, dr := range deploymentRequests {
		outputList = append(outputList, createDeploymentRequestJSONFromDbModel(dr))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// CreateDeploymentRequest ...
func (ctx Context) CreateDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	if !AuthorizeApplicationAction(ginctx, orgMember, application, ActionCreateDeploymentRequest) {
		return
	}

	var input deploymentRequestJSON
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	deploymentRequest := dbmodels.DeploymentRequest{
		BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
		ApplicationID: applicationID,
		State:         deploymentrequeststate.Approved, // TODO: set to InProgress when developing rules engine
	}
	patchDeploymentRequestDbModelFromJSON(&deploymentRequest, input)
	if err = ctx.Db.Create(&deploymentRequest).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest)
	ginctx.JSON(http.StatusOK, output)
}

// GetDeploymentRequest ...
func (ctx Context) GetDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	deploymentRequestID := ginctx.Param("id")
	applicationID := ginctx.Param("application_id")

	deploymentRequest, err := dbmodels.FindDeploymentRequest(ctx.Db, orgID, applicationID, deploymentRequestID)
	if err != nil {
		respondWithDbQueryError("deployment request", err, ginctx)
		return
	}

	if !AuthorizeDeploymentRequestAction(ginctx, orgMember, deploymentRequest, ActionReadDeploymentRequest) {
		return
	}

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest)
	ginctx.JSON(http.StatusOK, output)
}

// PatchDeploymentRequest ...
func (ctx Context) PatchDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	deploymentRequestID := ginctx.Param("id")
	applicationID := ginctx.Param("application_id")

	deploymentRequest, err := dbmodels.FindDeploymentRequest(ctx.Db, orgID, applicationID, deploymentRequestID)
	if err != nil {
		respondWithDbQueryError("deployment request", err, ginctx)
		return
	}

	if !AuthorizeDeploymentRequestAction(ginctx, orgMember, deploymentRequest, ActionUpdateDeploymentRequest) {
		return
	}

	var input deploymentRequestJSON
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	patchDeploymentRequestDbModelFromJSON(&deploymentRequest, input)
	if err = ctx.Db.Save(&deploymentRequest).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest)
	ginctx.JSON(http.StatusOK, output)
}

// DeleteDeploymentRequest ...
func (ctx Context) DeleteDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	deploymentRequestID := ginctx.Param("id")
	applicationID := ginctx.Param("application_id")

	deploymentRequest, err := dbmodels.FindDeploymentRequest(ctx.Db, orgID, applicationID, deploymentRequestID)
	if err != nil {
		respondWithDbQueryError("deployment request", err, ginctx)
		return
	}

	if !AuthorizeDeploymentRequestAction(ginctx, orgMember, deploymentRequest, ActionDeleteDeploymentRequest) {
		return
	}

	if err = ctx.Db.Delete(&deploymentRequest).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest)
	ginctx.JSON(http.StatusOK, output)
}

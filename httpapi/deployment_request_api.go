package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAllDeploymentRequests ...
func (ctx Context) GetAllDeploymentRequests(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	if len(applicationID) > 0 {
		application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
		if err != nil {
			respondWithDbQueryError("application", err, ginctx)
			return
		}

		if !AuthorizeApplicationAction(ginctx, orgMember, application, ActionReadApplication) {
			return
		}
	} else if !AuthorizeDeploymentRequestAction(ginctx, orgMember, dbmodels.DeploymentRequest{},
		ActionReadAllDeploymentRequests) {

		return
	}

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deploymentRequests, err := dbmodels.FindAllDeploymentRequests(
		tx.Preload("Application").Order("created_at DESC"),
		orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("deployment requests", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectDeploymentRequestApplications(deploymentRequests))
	if err != nil {
		respondWithDbQueryError("applications", err, ginctx)
		return
	}

	outputList := make([]deploymentRequestJSON, 0, len(deploymentRequests))
	for _, dr := range deploymentRequests {
		outputList = append(outputList, createDeploymentRequestJSONFromDbModel(dr, len(applicationID) == 0))
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

	var deploymentRequest dbmodels.DeploymentRequest
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		deploymentRequest = dbmodels.DeploymentRequest{
			BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
			ApplicationID: applicationID,
			State:         deploymentrequeststate.Approved, // TODO: set to InProgress when developing rules engine
			FinalizedAt:   sql.NullTime{Time: time.Now(), Valid: true},
		}
		patchDeploymentRequestDbModelFromJSON(&deploymentRequest, input)
		if err := tx.Create(&deploymentRequest).Error; err != nil {
			return err
		}

		err = tx.Create(dbmodels.DeploymentRequestCreatedEvent{
			DeploymentRequestEvent: dbmodels.DeploymentRequestEvent{
				BaseModel:           dbmodels.BaseModel{OrganizationID: orgID},
				DeploymentRequestID: deploymentRequest.ID,
				ApplicationID:       applicationID,
			},
		}).Error
		if err != nil {
			return err
		}

		err = tx.Create(dbmodels.DeploymentRequestRuleProcessedEvent{
			DeploymentRequestEvent: dbmodels.DeploymentRequestEvent{
				BaseModel:           dbmodels.BaseModel{OrganizationID: orgID},
				DeploymentRequestID: deploymentRequest.ID,
				ApplicationID:       applicationID,
			},
			ResultState: deploymentrequeststate.Approved,
		}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

// GetDeploymentRequest ...
func (ctx Context) GetDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	deploymentRequestID, err := strconv.ParseUint(ginctx.Param("id"), 10, 64)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'id' parameter as an integer: " + err.Error()})
		return
	}

	deploymentRequest, err := dbmodels.FindDeploymentRequest(
		ctx.Db.Preload("Application"),
		orgID, applicationID, deploymentRequestID)
	if err != nil {
		respondWithDbQueryError("deployment request", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		[]*dbmodels.Application{&deploymentRequest.Application})
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	if !AuthorizeDeploymentRequestAction(ginctx, orgMember, deploymentRequest, ActionReadDeploymentRequest) {
		return
	}

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

// PatchDeploymentRequest ...
func (ctx Context) PatchDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	deploymentRequestID, err := strconv.ParseUint(ginctx.Param("id"), 10, 64)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'id' parameter as an integer: " + err.Error()})
		return
	}

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

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

// DeleteDeploymentRequest ...
func (ctx Context) DeleteDeploymentRequest(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	deploymentRequestID, err := strconv.ParseUint(ginctx.Param("id"), 10, 64)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'id' parameter as an integer: " + err.Error()})
		return
	}

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

	output := createDeploymentRequestJSONFromDbModel(deploymentRequest, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

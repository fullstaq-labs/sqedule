package httpapi

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetAllReleases ...
func (ctx Context) GetAllReleases(ginctx *gin.Context) {
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
	} else if !AuthorizeReleaseAction(ginctx, orgMember, dbmodels.Release{},
		ActionReadAllReleases) {

		return
	}

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	releases, err := dbmodels.FindAllReleases(
		tx.Preload("Application").Order("created_at DESC"),
		orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("releases", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectReleaseApplications(releases))
	if err != nil {
		respondWithDbQueryError("applications", err, ginctx)
		return
	}

	outputList := make([]releaseJSON, 0, len(releases))
	for _, dr := range releases {
		outputList = append(outputList, createReleaseJSONFromDbModel(dr, len(applicationID) == 0))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// CreateRelease ...
func (ctx Context) CreateRelease(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	if !AuthorizeApplicationAction(ginctx, orgMember, application, ActionCreateRelease) {
		return
	}

	var input releaseJSON
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var release dbmodels.Release
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		release = dbmodels.Release{
			BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
			ApplicationID: applicationID,
			State:         releasestate.Approved, // TODO: set to InProgress when developing rules engine
			FinalizedAt:   sql.NullTime{Time: time.Now(), Valid: true},
		}
		patchReleaseDbModelFromJSON(&release, input)
		if err := tx.Create(&release).Error; err != nil {
			return err
		}

		err = tx.Create(dbmodels.ReleaseCreatedEvent{
			ReleaseEvent: dbmodels.ReleaseEvent{
				BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
				ReleaseID:     release.ID,
				ApplicationID: applicationID,
			},
		}).Error
		if err != nil {
			return err
		}

		err = tx.Create(dbmodels.ReleaseRuleProcessedEvent{
			ReleaseEvent: dbmodels.ReleaseEvent{
				BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
				ReleaseID:     release.ID,
				ApplicationID: applicationID,
			},
			ResultState: releasestate.Approved,
		}).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	output := createReleaseJSONFromDbModel(release, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

// GetRelease ...
func (ctx Context) GetRelease(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	releaseID, err := strconv.ParseUint(ginctx.Param("id"), 10, 64)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'id' parameter as an integer: " + err.Error()})
		return
	}

	release, err := dbmodels.FindRelease(
		ctx.Db.Preload("Application"),
		orgID, applicationID, releaseID)
	if err != nil {
		respondWithDbQueryError("release", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		[]*dbmodels.Application{&release.Application})
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	if !AuthorizeReleaseAction(ginctx, orgMember, release, ActionReadRelease) {
		return
	}

	output := createReleaseJSONFromDbModel(release, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

// PatchRelease ...
func (ctx Context) PatchRelease(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	releaseID, err := strconv.ParseUint(ginctx.Param("id"), 10, 64)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'id' parameter as an integer: " + err.Error()})
		return
	}

	release, err := dbmodels.FindRelease(ctx.Db, orgID, applicationID, releaseID)
	if err != nil {
		respondWithDbQueryError("release", err, ginctx)
		return
	}

	if !AuthorizeReleaseAction(ginctx, orgMember, release, ActionUpdateRelease) {
		return
	}

	var input releaseJSON
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	patchReleaseDbModelFromJSON(&release, input)
	if err = ctx.Db.Save(&release).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createReleaseJSONFromDbModel(release, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

// DeleteRelease ...
func (ctx Context) DeleteRelease(ginctx *gin.Context) {
	orgMember := getAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")

	releaseID, err := strconv.ParseUint(ginctx.Param("id"), 10, 64)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'id' parameter as an integer: " + err.Error()})
		return
	}

	release, err := dbmodels.FindRelease(ctx.Db, orgID, applicationID, releaseID)
	if err != nil {
		respondWithDbQueryError("release", err, ginctx)
		return
	}

	if !AuthorizeReleaseAction(ginctx, orgMember, release, ActionDeleteRelease) {
		return
	}

	if err = ctx.Db.Delete(&release).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createReleaseJSONFromDbModel(release, len(applicationID) == 0)
	ginctx.JSON(http.StatusOK, output)
}

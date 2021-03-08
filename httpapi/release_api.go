package httpapi

import (
	"net/http"
	"strconv"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetAllReleases ...
func (ctx Context) GetAllReleases(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")
	includeAppJSON := len(applicationID) == 0

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
	if includeAppJSON {
		tx = tx.Preload("Application")
	}
	releases, err := dbmodels.FindAllReleases(
		tx.Order("created_at DESC"),
		orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("releases", err, ginctx)
		return
	}

	if includeAppJSON {
		err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
			dbmodels.CollectReleaseApplications(releases))
		if err != nil {
			respondWithDbQueryError("application versions", err, ginctx)
			return
		}
	}

	outputList := make([]releaseJSON, 0, len(releases))
	for _, release := range releases {
		outputList = append(outputList,
			createReleaseJSONFromDbModel(release, len(applicationID) == 0, nil))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// CreateRelease ...
func (ctx Context) CreateRelease(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")
	includeAppJSON := len(applicationID) == 0

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

	if includeAppJSON {
		err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID, []*dbmodels.Application{&application})
		if err != nil {
			respondWithDbQueryError("application versions", err, ginctx)
			return
		}
	}

	var release dbmodels.Release
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		release = dbmodels.Release{
			BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
			ApplicationID: applicationID,
			State:         releasestate.InProgress,
		}
		patchReleaseDbModelFromJSON(&release, input)
		if err := tx.Create(&release).Error; err != nil {
			return err
		}

		appRuleBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindings(ctx.Db, orgID, applicationID)
		if err != nil {
			return err
		}
		err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(ctx.Db, orgID,
			dbmodels.MakeApplicationApprovalRulesetBindingPointerArray(appRuleBindings))
		if err != nil {
			return err
		}
		err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
			dbmodels.CollectApplicationApprovalRulesetBindingRulesets(appRuleBindings))
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateReleaseApprovalRulesetBindings(ctx.Db, release.ID, appRuleBindings)
		if err != nil {
			return err
		}

		createdEvent := dbmodels.ReleaseCreatedEvent{
			ReleaseEvent: dbmodels.ReleaseEvent{
				BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
				ReleaseID:     release.ID,
				ApplicationID: applicationID,
			},
		}
		err = tx.Create(&createdEvent).Error
		if err != nil {
			return err
		}

		creationRecord := dbmodels.NewCreationAuditRecord(orgID, orgMember, ginctx.ClientIP())
		creationRecord.ReleaseCreatedEventID = &createdEvent.ID
		err = tx.Omit(clause.Associations).Create(&creationRecord).Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var bindings []dbmodels.ReleaseApprovalRulesetBinding
	output := createReleaseJSONFromDbModel(release, includeAppJSON, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

// GetRelease ...
func (ctx Context) GetRelease(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")
	includeAppJSON := len(applicationID) == 0

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

	if includeAppJSON {
		err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
			[]*dbmodels.Application{&release.Application})
		if err != nil {
			respondWithDbQueryError("application versions", err, ginctx)
			return
		}
	}

	if !AuthorizeReleaseAction(ginctx, orgMember, release, ActionReadRelease) {
		return
	}

	bindings, err := dbmodels.FindAllReleaseApprovalRulesetBindings(
		ctx.Db.Preload("ApprovalRuleset").
			Preload("ApprovalRulesetMajorVersion").
			Preload("ApprovalRulesetMinorVersion"),
		orgID, applicationID, release.ID)
	if err != nil {
		respondWithDbQueryError("release approval ruleset bindings", err, ginctx)
		return
	}

	output := createReleaseJSONFromDbModel(release, includeAppJSON, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

// PatchRelease ...
func (ctx Context) PatchRelease(ginctx *gin.Context) {
	orgMember := GetAuthenticatedOrganizationMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationMember().BaseModel.OrganizationID
	applicationID := ginctx.Param("application_id")
	includeAppJSON := len(applicationID) == 0

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

	if includeAppJSON {
		err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID, []*dbmodels.Application{&release.Application})
		if err != nil {
			respondWithDbQueryError("application versions", err, ginctx)
			return
		}
	}

	bindings, err := dbmodels.FindAllReleaseApprovalRulesetBindings(
		ctx.Db.Preload("ApprovalRuleset").
			Preload("ApprovalRulesetMajorVersion").
			Preload("ApprovalRulesetMinorVersion"),
		orgID, applicationID, release.ID)
	if err != nil {
		respondWithDbQueryError("release approval ruleset bindings", err, ginctx)
		return
	}

	patchReleaseDbModelFromJSON(&release, input)
	if err = ctx.Db.Save(&release).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := createReleaseJSONFromDbModel(release, includeAppJSON, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

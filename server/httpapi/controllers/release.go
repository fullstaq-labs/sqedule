package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fullstaq-labs/sqedule/server/approvalrulesprocessing"
	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetAllReleases ...
func (ctx Context) GetAllReleases(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	includeAppJSON := len(applicationID) == 0

	if len(applicationID) > 0 {
		application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
		if err != nil {
			respondWithDbQueryError("application", err, ginctx)
			return
		}

		authorizer := authz.ApplicationAuthorizer{}
		if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, application) {
			respondWithUnauthorizedError(ginctx)
			return
		}
	} else if !authz.AuthorizeCollectionAction(authz.ReleaseAuthorizer{}, orgMember, authz.ActionListReleases) {
		respondWithUnauthorizedError(ginctx)
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
			dbmodels.CollectApplicationsWithReleases(dbmodels.MakeReleasesPointerArray(releases)))
		if err != nil {
			respondWithDbQueryError("application versions", err, ginctx)
			return
		}
	}

	outputList := make([]json.ReleaseWithAssociations, 0, len(releases))
	for _, release := range releases {
		outputList = append(outputList,
			json.CreateFromDbReleaseWithAssociations(release, len(applicationID) == 0, nil))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

// CreateRelease ...
func (ctx Context) CreateRelease(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	includeAppJSON := len(applicationID) == 0

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionCreateRelease, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	var input json.ReleasePatchablePart
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
	var releaseRulesetBindings []dbmodels.ReleaseApprovalRulesetBinding
	var job dbmodels.ReleaseBackgroundJob
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		release = dbmodels.Release{
			BaseModel:     dbmodels.BaseModel{OrganizationID: orgID},
			ApplicationID: applicationID,
			State:         releasestate.InProgress,
		}
		json.PatchDbRelease(&release, input)
		if err := tx.Create(&release).Error; err != nil {
			return err
		}

		appRulesetBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindings(
			tx.Preload("ApprovalRuleset"), orgID, applicationID)
		if err != nil {
			return err
		}
		err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(tx, orgID,
			dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appRulesetBindings))
		if err != nil {
			return err
		}
		err = dbmodels.LoadApprovalRulesetsLatestVersions(tx, orgID,
			dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(appRulesetBindings))
		if err != nil {
			return err
		}

		releaseRulesetBindings, err = dbmodels.CreateReleaseApprovalRulesetBindings(tx, release.ID, appRulesetBindings)
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

		job, err = dbmodels.CreateReleaseBackgroundJob(tx, orgID, applicationID, release)
		if err != nil {
			return fmt.Errorf("Error creating background job for processing this Release: %w", err)
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = approvalrulesprocessing.ProcessInBackground(ctx.Db, orgID, job)
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	output := json.CreateFromDbReleaseWithAssociations(release, includeAppJSON, &releaseRulesetBindings)
	ginctx.JSON(http.StatusOK, output)
}

// GetRelease ...
func (ctx Context) GetRelease(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
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

	authorizer := authz.ReleaseAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadRelease, release) {
		respondWithUnauthorizedError(ginctx)
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

	output := json.CreateFromDbReleaseWithAssociations(release, includeAppJSON, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

// PatchRelease ...
func (ctx Context) PatchRelease(ginctx *gin.Context) {
	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
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

	authorizer := authz.ReleaseAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateRelease, release) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	var input json.ReleasePatchablePart
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

	json.PatchDbRelease(&release, input)
	if err = ctx.Db.Save(&release).Error; err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	output := json.CreateFromDbReleaseWithAssociations(release, includeAppJSON, &bindings)
	ginctx.JSON(http.StatusOK, output)
}

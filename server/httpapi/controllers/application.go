package controllers

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//
// ******** Operations on resources ********
//

func (ctx Context) CreateApplication(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	var input json.ApplicationInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	if input.ID == nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: 'id' field must be set"})
		return
	}
	if input.Version == nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: 'version' field must be set"})
		return
	}
	if !input.Version.ProposalState.IsEffectivelyDraft() && input.Version.ProposalState != proposalstate.Final {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: version.proposal_state must be either draft or final ('" +
			input.Version.ProposalState + "' given)"})
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionCreateApplication) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Modify database

	app := dbmodels.Application{
		BaseModel: dbmodels.BaseModel{OrganizationID: orgID},
		ID:        *input.ID,
	}
	version, adjustment := app.NewDraftVersion()
	app.Version = version
	json.PatchApplication(&app, input)
	json.PatchApplicationAdjustment(orgID, adjustment, *input.Version)
	if input.Version.ProposalState == proposalstate.Final {
		dbmodels.FinalizeReviewableProposal(&version.ReviewableVersionBase,
			&adjustment.ReviewableAdjustmentBase, 0,
			app.CheckNewProposalsRequireReview(dbmodels.ReviewableActionCreate))
	}

	err := ctx.Db.Transaction(func(tx *gorm.DB) error {
		err := tx.Omit(clause.Associations).Create(&app).Error
		if err != nil {
			return err
		}

		err = tx.Omit(clause.Associations).Create(version).Error
		if err != nil {
			return err
		}

		adjustment.ApplicationVersionID = version.ID
		err = tx.Omit(clause.Associations).Create(adjustment).Error
		if err != nil {
			return err
		}

		creationRecord := dbmodels.NewCreationAuditRecord(orgID, nil, "")
		creationRecord.ApplicationVersionID = &version.ID
		creationRecord.ApplicationAdjustmentNumber = &adjustment.AdjustmentNumber
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

	// Generate response

	output := json.CreateApplicationWithVersionAndAssociations(app, version, nil)
	ginctx.JSON(http.StatusCreated, output)
}

func (ctx Context) ListApplications(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionListApplications) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	apps, err := dbmodels.FindApplications(
		tx.Order("created_at DESC"),
		orgID)
	if err != nil {
		respondWithDbQueryError("applications", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, dbmodels.MakeApplicationsPointerArray(apps))
	if err != nil {
		respondWithDbQueryError("application versions", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApplicationWithLatestApprovedVersion, 0, len(apps))
	for _, app := range apps {
		outputList = append(outputList, json.CreateApplicationWithLatestApprovedVersion(app, app.Version))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApplication(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.Application{&app})
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, app) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	bindings, err := dbmodels.FindApplicationApprovalRulesetBindings(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(bindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings))
	if err != nil {
		respondWithDbQueryError("approval ruleset versions", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateApplicationWithLatestApprovedVersionAndRulesetBindings(app, app.Version, bindings)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateApplication(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	var input json.ApplicationInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateApplication, app) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.Application{&app})
	if err != nil {
		respondWithDbQueryError("application latest version", err, ginctx)
		return
	}

	rulesetBindings, err := dbmodels.FindApplicationApprovalRulesetBindingsWithApplication(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(rulesetBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings latest versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID, dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(rulesetBindings))
	if err != nil {
		respondWithDbQueryError("approval rulesets latest versions", err, ginctx)
		return
	}

	var latestApprovedVersionNumber uint32 = 0

	if app.Version != nil {
		latestApprovedVersionNumber = *app.Version.VersionNumber
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		var appUpdate dbmodels.Application = app
		json.PatchApplication(&appUpdate, input)
		savetx := tx.Omit(clause.Associations).Model(&app).Updates(appUpdate)
		if savetx.Error != nil {
			return savetx.Error
		}

		if input.Version != nil {
			newVersion, newAdjustment := app.NewDraftVersion()
			json.PatchApplicationAdjustment(orgID, newAdjustment, *input.Version)

			if input.Version.ProposalState == proposalstate.Final {
				dbmodels.FinalizeReviewableProposal(&newVersion.ReviewableVersionBase,
					&newAdjustment.ReviewableAdjustmentBase,
					latestApprovedVersionNumber,
					app.CheckNewProposalsRequireReview(dbmodels.ReviewableActionUpdate))
			} else {
				dbmodels.SetReviewableAdjustmentReviewStateFromUnfinalizedProposalState(&newAdjustment.ReviewableAdjustmentBase,
					input.Version.ProposalState)
			}
			if err = tx.Omit(clause.Associations).Create(newVersion).Error; err != nil {
				return err
			}

			newAdjustment.ApplicationVersionID = newVersion.ID
			if err = tx.Omit(clause.Associations).Create(newAdjustment).Error; err != nil {
				return err
			}

			creationRecord := dbmodels.NewCreationAuditRecord(orgID, orgMember, ginctx.ClientIP())
			creationRecord.ApplicationVersionID = &newVersion.ID
			creationRecord.ApplicationAdjustmentNumber = &newAdjustment.AdjustmentNumber
			err = tx.Omit(clause.Associations).Create(&creationRecord).Error
			if err != nil {
				return err
			}

			app.Version = newVersion
			app.Version.Adjustment = newAdjustment
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateApplicationWithVersionAndAssociations(app, app.Version, &rulesetBindings)
	ginctx.JSON(http.StatusOK, output)
}

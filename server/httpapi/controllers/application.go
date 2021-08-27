package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/proposalstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstateinput"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/reviewstateinput"
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
	if !input.Version.ProposalState.IsEffectivelyDraft() && input.Version.ProposalState != proposalstateinput.Final {
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
	if input.Version.ProposalState == proposalstateinput.Final {
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

			if input.Version.ProposalState == proposalstateinput.Final {
				dbmodels.FinalizeReviewableProposal(&newVersion.ReviewableVersionBase,
					&newAdjustment.ReviewableAdjustmentBase,
					latestApprovedVersionNumber,
					app.CheckNewProposalsRequireReview(dbmodels.ReviewableActionUpdate))
			} else {
				dbmodels.SetReviewableAdjustmentProposalStateFromProposalStateInput(&newAdjustment.ReviewableAdjustmentBase,
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

//
// ******** Operations on approved versions ********
//

func (ctx Context) ListApplicationVersions(ginctx *gin.Context) {
	ctx.listApplicationVersionsOrProposals(ginctx, true)
}

func (ctx Context) listApplicationVersionsOrProposals(ginctx *gin.Context, approved bool) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
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

	pagination, err := dbutils.ParsePaginationOptions(ginctx)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	versions, err := dbmodels.FindApplicationVersions(ctx.Db, orgID, id, approved, pagination)
	if err != nil {
		respondWithDbQueryError("application versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationVersionsPointerArray(versions))
	if err != nil {
		respondWithDbQueryError("application adjustments", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApplicationVersion, 0, len(versions))
	for _, version := range versions {
		outputList = append(outputList, json.CreateApplicationVersion(version))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApplicationVersion(ginctx *gin.Context) {
	ctx.getApplicationVersionOrProposal(ginctx, true)
}

func (ctx Context) getApplicationVersionOrProposal(ginctx *gin.Context, approved bool) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")

	var versionNumberOrID uint64
	var err error

	if approved {
		versionNumberOrID, err = strconv.ParseUint(ginctx.Param("version_number"), 10, 32)
		if err != nil {
			ginctx.JSON(http.StatusBadRequest,
				gin.H{"error": "Error parsing 'version_number' parameter as an integer: " + err.Error()})
			return
		}
	} else {
		versionNumberOrID, err = strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
		if err != nil {
			ginctx.JSON(http.StatusBadRequest,
				gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
			return
		}
	}

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
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

	binding, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	var version dbmodels.ApplicationVersion
	if approved {
		version, err = dbmodels.FindApplicationVersionByNumber(ctx.Db, orgID, id, uint32(versionNumberOrID))
		if err != nil {
			respondWithDbQueryError("application version", err, ginctx)
			return
		}
	} else {
		version, err = dbmodels.FindApplicationProposalByID(ctx.Db, orgID, id, versionNumberOrID)
		if err != nil {
			respondWithDbQueryError("application proposal", err, ginctx)
			return
		}
	}
	binding.Version = &version

	err = dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApplicationVersion{&version})
	if err != nil {
		respondWithDbQueryError("application adjustment", err, ginctx)
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

	// Generate response
	output := json.CreateApplicationWithVersionAndAssociations(binding, binding.Version, &rulesetBindings)
	ginctx.JSON(http.StatusOK, output)
}

//
// ******** Operations on proposals ********
//

func (ctx Context) ListApplicationProposals(ginctx *gin.Context) {
	ctx.listApplicationVersionsOrProposals(ginctx, false)
}

func (ctx Context) GetApplicationProposal(ginctx *gin.Context) {
	ctx.getApplicationVersionOrProposal(ginctx, false)
}

func (ctx Context) UpdateApplicationProposal(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	var input json.ApplicationVersionInput
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

	var latestApprovedVersionNumber uint32 = 0

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.Application{&app})
	if err != nil {
		respondWithDbQueryError("application latest approved version", err, ginctx)
		return
	}
	if app.Version != nil {
		latestApprovedVersionNumber = *app.Version.VersionNumber
	}

	proposals, err := dbmodels.FindApplicationProposals(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application proposals", err, ginctx)
		return
	}

	proposal := dbmodels.CollectApplicationVersionIDEquals(proposals, versionID)
	if proposal == nil {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": "application proposal not found"})
		return
	}

	otherProposals := dbmodels.CollectApplicationVersionIDNotEquals(proposals, versionID)

	err = dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationVersionsPointerArray(proposals))
	if err != nil {
		respondWithDbQueryError("application adjustment", err, ginctx)
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

	if proposal.Adjustment.ProposalState == proposalstate.Reviewing && input.ProposalState == proposalstateinput.Final {
		ginctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Cannot finalize a proposal which is already being reviewed"})
		return
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		// Create new Adjustment with patched state

		newAdjustment := proposal.Adjustment.NewAdjustment()
		json.PatchApplicationAdjustment(orgID, &newAdjustment, input)

		if input.ProposalState == proposalstateinput.Final {
			proposalUpdate := proposal
			dbmodels.FinalizeReviewableProposal(&proposalUpdate.ReviewableVersionBase,
				&newAdjustment.ReviewableAdjustmentBase,
				latestApprovedVersionNumber,
				app.CheckNewProposalsRequireReview(dbmodels.ReviewableActionCreate))
			if err = tx.Omit(clause.Associations).Model(&proposal).Updates(proposalUpdate).Error; err != nil {
				return err
			}
		} else {
			dbmodels.SetReviewableAdjustmentProposalStateFromProposalStateInput(&newAdjustment.ReviewableAdjustmentBase,
				input.ProposalState)
		}

		err = tx.Omit(clause.Associations).Create(&newAdjustment).Error
		if err != nil {
			return err
		}

		creationRecord := dbmodels.NewCreationAuditRecord(orgID, orgMember, ginctx.ClientIP())
		creationRecord.ApplicationVersionID = &proposal.ID
		creationRecord.ApplicationAdjustmentNumber = &newAdjustment.AdjustmentNumber
		err = tx.Omit(clause.Associations).Create(&creationRecord).Error
		if err != nil {
			return err
		}

		proposal.Adjustment = &newAdjustment

		if newAdjustment.ProposalState == proposalstate.Approved {
			// Ensure no other proposals are in the Reviewing state

			for _, proposal := range otherProposals {
				if proposal.Adjustment.ProposalState != proposalstate.Reviewing {
					continue
				}

				newAdjustment := proposal.Adjustment.NewAdjustment()
				newAdjustment.ProposalState = proposalstate.Draft
				err = tx.Omit(clause.Associations).Create(&newAdjustment).Error
				if err != nil {
					return err
				}

				creationRecord := dbmodels.NewCreationAuditRecord(orgID, nil, "")
				creationRecord.ApplicationVersionID = &proposal.ID
				creationRecord.ApplicationAdjustmentNumber = &newAdjustment.AdjustmentNumber
				err = tx.Omit(clause.Associations).Create(&creationRecord).Error
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateApplicationWithVersionAndAssociations(app, proposal, &rulesetBindings)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateApplicationProposalState(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	var input json.ReviewableProposalStateInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReviewApplication, app) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	var latestApprovedVersionNumber uint32 = 0

	if input.State == reviewstateinput.Approved {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.Application{&app})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest approved version", err, ginctx)
			return
		}

		if app.Version != nil {
			latestApprovedVersionNumber = *app.Version.VersionNumber
		}
	}

	proposals, err := dbmodels.FindApplicationProposals(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application proposals", err, ginctx)
		return
	}

	proposal := dbmodels.CollectApplicationVersionIDEquals(proposals, versionID)
	if proposal == nil {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": "application proposal not found"})
		return
	}

	otherProposals := dbmodels.CollectApplicationVersionIDNotEquals(proposals, versionID)

	err = dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationVersionsPointerArray(proposals))
	if err != nil {
		respondWithDbQueryError("application adjustments", err, ginctx)
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

	if proposal.Adjustment.ProposalState != proposalstate.Reviewing {
		ginctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "This proposal is not awaiting review"})
		return
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		// Create new Adjustment with new review state

		proposalUpdate := *proposal
		newAdjustment := proposal.Adjustment.NewAdjustment()

		if input.State == reviewstateinput.Approved {
			newAdjustment.ProposalState = proposalstate.Approved
			proposalUpdate.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}
			proposalUpdate.VersionNumber = lib.NewUint32Ptr(latestApprovedVersionNumber + 1)
		} else {
			newAdjustment.ProposalState = proposalstate.Rejected
		}

		if err = tx.Omit(clause.Associations).Model(proposal).Updates(proposalUpdate).Error; err != nil {
			return err
		}
		if err = tx.Omit(clause.Associations).Create(&newAdjustment).Error; err != nil {
			return err
		}

		creationRecord := dbmodels.NewCreationAuditRecord(orgID, orgMember, ginctx.ClientIP())
		creationRecord.ApplicationVersionID = &proposal.ID
		creationRecord.ApplicationAdjustmentNumber = &newAdjustment.AdjustmentNumber
		err = tx.Omit(clause.Associations).Create(&creationRecord).Error
		if err != nil {
			return err
		}

		proposal.Adjustment = &newAdjustment

		if input.State == reviewstateinput.Approved {
			// Ensure no other proposals are in the Reviewing state

			for _, proposal := range otherProposals {
				if proposal.Adjustment.ProposalState != proposalstate.Reviewing {
					continue
				}

				newAdjustment := proposal.Adjustment.NewAdjustment()
				newAdjustment.ProposalState = proposalstate.Draft
				err = tx.Omit(clause.Associations).Create(&newAdjustment).Error
				if err != nil {
					return err
				}

				creationRecord := dbmodels.NewCreationAuditRecord(orgID, nil, "")
				creationRecord.ApplicationVersionID = &proposal.ID
				creationRecord.ApplicationAdjustmentNumber = &newAdjustment.AdjustmentNumber
				err = tx.Omit(clause.Associations).Create(&creationRecord).Error
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateApplicationWithVersionAndAssociations(app, proposal, &rulesetBindings)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) DeleteApplicationProposal(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("application_id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	app, err := dbmodels.FindApplication(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionDeleteApplication, app) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	version, err := dbmodels.FindApplicationProposalByID(ctx.Db, orgID, id, versionID)
	if err != nil {
		respondWithDbQueryError("application proposal", err, ginctx)
		return
	}
	app.Version = &version

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		err = dbmodels.DeleteAuditCreationRecordsForApplicationProposal(tx, orgID, version.ID)
		if err != nil {
			return err
		}
		err = dbmodels.DeleteApplicationAdjustmentsForProposal(tx, orgID, version.ID)
		if err != nil {
			return err
		}

		err = tx.Delete(&version).Error
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

	ginctx.JSON(http.StatusOK, gin.H{})
}

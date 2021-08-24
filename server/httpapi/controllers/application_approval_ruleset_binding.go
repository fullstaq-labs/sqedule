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

func (ctx Context) CreateApplicationApprovalRulesetBinding(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ApplicationApprovalRulesetBindingInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
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

	appAuthorizer := authz.ApplicationAuthorizer{}
	appProposeBindAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionProposeBindApplicationToApprovalRuleset, application)
	appReadAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetProposeBindAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionProposeBindApprovalRulesetToApplication, ruleset)
	rulesetReadAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appProposeBindAuthorized || !rulesetProposeBindAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Modify database

	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.Application{&application})
	if err != nil {
		respondWithDbQueryError("application latest version", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest version", err, ginctx)
		return
	}

	binding := dbmodels.ApplicationApprovalRulesetBinding{
		BaseModel: dbmodels.BaseModel{OrganizationID: orgID},
		ApplicationApprovalRulesetBindingPrimaryKey: dbmodels.ApplicationApprovalRulesetBindingPrimaryKey{
			ApplicationID:     applicationID,
			ApprovalRulesetID: rulesetID,
		},
		Application:     application,
		ApprovalRuleset: ruleset,
	}
	version, adjustment := binding.NewDraftVersion()
	binding.Version = version
	json.PatchApplicationApprovalRulesetBinding(&binding, input)
	json.PatchApplicationApprovalRulesetBindingAdjustment(orgID, adjustment, *input.Version)
	if input.Version.ProposalState == proposalstateinput.Final {
		dbmodels.FinalizeReviewableProposal(&version.ReviewableVersionBase,
			&adjustment.ReviewableAdjustmentBase, 0,
			binding.CheckNewProposalsRequireReview(
				dbmodels.ReviewableActionCreate,
				version.Adjustment.Mode))
	}

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		err := tx.Omit(clause.Associations).Create(&binding).Error
		if err != nil {
			return err
		}

		err = tx.Omit(clause.Associations).Create(version).Error
		if err != nil {
			return err
		}

		adjustment.ApplicationApprovalRulesetBindingVersionID = version.ID
		err = tx.Omit(clause.Associations).Create(adjustment).Error
		if err != nil {
			return err
		}

		creationRecord := dbmodels.NewCreationAuditRecord(orgID, nil, "")
		creationRecord.ApplicationApprovalRulesetBindingVersionID = &version.ID
		creationRecord.ApplicationApprovalRulesetBindingAdjustmentNumber = &adjustment.AdjustmentNumber
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

	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, version,
		appReadAuthorized, rulesetReadAuthorized)
	ginctx.JSON(http.StatusCreated, output)
}

func (ctx Context) ListApplicationApprovalRulesetBindings(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")

	// Check authorization

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
	} else {
		authorizer := authz.ApplicationApprovalRulesetBindingAuthorizer{}
		if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionListApplicationApprovalRulesetBindings) {
			respondWithUnauthorizedError(ginctx)
			return
		}
	}

	// TODO: filter out bound apps and rulesets that the client is not authorized to read

	// Query database

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx = tx.Preload("ApprovalRuleset")
	if len(applicationID) == 0 {
		tx = tx.Preload("Application")
	}
	bindings, err := dbmodels.FindApplicationApprovalRulesetBindings(tx, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(bindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithApplicationApprovalRulesetBindings(bindings))
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	if len(applicationID) == 0 {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
			dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(bindings))
		if err != nil {
			respondWithDbQueryError("application latest versions", err, ginctx)
			return
		}
	}

	// Generate response

	outputList := make([]json.ApplicationApprovalRulesetBindingWithLatestApprovedVersion, 0, len(bindings))
	for _, binding := range bindings {
		outputList = append(outputList,
			json.CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(
				binding, binding.Version, len(applicationID) == 0, true))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApplicationApprovalRulesetBinding(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appAuthorized && !rulesetAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApplicationApprovalRulesetBinding{&binding})
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest version", err, ginctx)
		return
	}

	if appAuthorized {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.Application{&application})
		if err != nil {
			respondWithDbQueryError("application latest version", err, ginctx)
			return
		}

		binding.Application = application
	}

	if rulesetAuthorized {
		err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.ApprovalRuleset{&ruleset})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest version", err, ginctx)
			return
		}

		binding.ApprovalRuleset = ruleset
	}

	// Generate response

	output := json.CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(binding, binding.Version,
		appAuthorized, rulesetAuthorized)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateApplicationApprovalRulesetBinding(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ApplicationApprovalRulesetBindingInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appProposeBindAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionProposeBindApplicationToApprovalRuleset, application)
	appReadAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetProposeBindAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionProposeBindApprovalRulesetToApplication, ruleset)
	rulesetReadAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appProposeBindAuthorized || !rulesetProposeBindAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApplicationApprovalRulesetBinding{&binding})
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest version", err, ginctx)
		return
	}

	if appReadAuthorized {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.Application{&application})
		if err != nil {
			respondWithDbQueryError("application latest version", err, ginctx)
			return
		}

		binding.Application = application
	}

	if rulesetReadAuthorized {
		err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.ApprovalRuleset{&ruleset})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest version", err, ginctx)
			return
		}

		binding.ApprovalRuleset = ruleset
	}

	var latestApprovedVersionNumber uint32 = 0

	if binding.Version != nil {
		latestApprovedVersionNumber = *binding.Version.VersionNumber
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		var bindingUpdate dbmodels.ApplicationApprovalRulesetBinding = binding
		json.PatchApplicationApprovalRulesetBinding(&bindingUpdate, input)
		savetx := tx.Omit(clause.Associations).Model(&binding).Updates(bindingUpdate)
		if savetx.Error != nil {
			return savetx.Error
		}

		if input.Version != nil {
			newVersion, newAdjustment := binding.NewDraftVersion()
			json.PatchApplicationApprovalRulesetBindingAdjustment(orgID, newAdjustment, *input.Version)

			if input.Version.ProposalState == proposalstateinput.Final {
				dbmodels.FinalizeReviewableProposal(&newVersion.ReviewableVersionBase,
					&newAdjustment.ReviewableAdjustmentBase,
					latestApprovedVersionNumber,
					binding.CheckNewProposalsRequireReview(
						dbmodels.ReviewableActionUpdate,
						newAdjustment.Mode))
			} else {
				dbmodels.SetReviewableAdjustmentProposalStateFromProposalStateInput(&newAdjustment.ReviewableAdjustmentBase,
					input.Version.ProposalState)
			}
			if err = tx.Omit(clause.Associations).Create(newVersion).Error; err != nil {
				return err
			}

			newAdjustment.ApplicationApprovalRulesetBindingVersionID = newVersion.ID
			if err = tx.Omit(clause.Associations).Create(newAdjustment).Error; err != nil {
				return err
			}

			creationRecord := dbmodels.NewCreationAuditRecord(orgID, orgMember, ginctx.ClientIP())
			creationRecord.ApplicationApprovalRulesetBindingVersionID = &newVersion.ID
			creationRecord.ApplicationApprovalRulesetBindingAdjustmentNumber = &newAdjustment.AdjustmentNumber
			err = tx.Omit(clause.Associations).Create(&creationRecord).Error
			if err != nil {
				return err
			}

			binding.Version = newVersion
			binding.Version.Adjustment = newAdjustment
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, binding.Version,
		appReadAuthorized, rulesetReadAuthorized)
	ginctx.JSON(http.StatusOK, output)
}

//
// ******** Operations on approved versions ********
//

func (ctx Context) ListApplicationApprovalRulesetBindingVersions(ginctx *gin.Context) {
	ctx.listApplicationApprovalRulesetBindingVersionsOrProposals(ginctx, true)
}

func (ctx Context) listApplicationApprovalRulesetBindingVersionsOrProposals(ginctx *gin.Context, approved bool) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appAuthorized || !rulesetAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	pagination, err := dbutils.ParsePaginationOptions(ginctx)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	versions, err := dbmodels.FindApplicationApprovalRulesetBindingVersions(ctx.Db, orgID, applicationID, rulesetID, approved, pagination)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingVersionsPointerArray(versions))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding adjustments", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApplicationApprovalRulesetBindingVersion, 0, len(versions))
	for _, version := range versions {
		outputList = append(outputList, json.CreateApplicationApprovalRulesetBindingVersion(version))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApplicationApprovalRulesetBindingVersion(ginctx *gin.Context) {
	ctx.getApplicationApprovalRulesetBindingVersionOrProposal(ginctx, true)
}

func (ctx Context) getApplicationApprovalRulesetBindingVersionOrProposal(ginctx *gin.Context, approved bool) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")

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

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appAuthorized || !rulesetAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	if appAuthorized {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.Application{&application})
		if err != nil {
			respondWithDbQueryError("application latest version", err, ginctx)
			return
		}

		binding.Application = application
	}

	if rulesetAuthorized {
		err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&ruleset})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest version", err, ginctx)
			return
		}

		binding.ApprovalRuleset = ruleset
	}

	var version dbmodels.ApplicationApprovalRulesetBindingVersion
	if approved {
		version, err = dbmodels.FindApplicationApprovalRulesetBindingVersionByNumber(ctx.Db, orgID, applicationID, rulesetID, uint32(versionNumberOrID))
		if err != nil {
			respondWithDbQueryError("application approval ruleset binding version", err, ginctx)
			return
		}
	} else {
		version, err = dbmodels.FindApplicationApprovalRulesetBindingProposalByID(ctx.Db, orgID, applicationID, rulesetID, versionNumberOrID)
		if err != nil {
			respondWithDbQueryError("application approval ruleset binding proposal", err, ginctx)
			return
		}
	}
	binding.Version = &version

	err = dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApplicationApprovalRulesetBindingVersion{&version})
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding adjustment", err, ginctx)
		return
	}

	// Generate response
	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, binding.Version,
		appAuthorized, rulesetAuthorized)
	ginctx.JSON(http.StatusOK, output)
}

//
// ******** Operations on proposals ********
//

func (ctx Context) ListApplicationApprovalRulesetBindingProposals(ginctx *gin.Context) {
	ctx.listApplicationApprovalRulesetBindingVersionsOrProposals(ginctx, false)
}

func (ctx Context) GetApplicationApprovalRulesetBindingProposal(ginctx *gin.Context) {
	ctx.getApplicationApprovalRulesetBindingVersionOrProposal(ginctx, false)
}

func (ctx Context) UpdateApplicationApprovalRulesetBindingProposal(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ApplicationApprovalRulesetBindingVersionInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appProposeBindAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionProposeBindApplicationToApprovalRuleset, application)
	appReadAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetPrposeBindAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionProposeBindApprovalRulesetToApplication, ruleset)
	rulesetReadAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appProposeBindAuthorized || !rulesetPrposeBindAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	var latestApprovedVersionNumber uint32 = 0

	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApplicationApprovalRulesetBinding{&binding})
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest approved version", err, ginctx)
		return
	}
	if binding.Version != nil {
		latestApprovedVersionNumber = *binding.Version.VersionNumber
	}

	if appReadAuthorized {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.Application{&application})
		if err != nil {
			respondWithDbQueryError("application latest version", err, ginctx)
			return
		}

		binding.Application = application
	}

	if rulesetReadAuthorized {
		err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.ApprovalRuleset{&ruleset})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest version", err, ginctx)
			return
		}

		binding.ApprovalRuleset = ruleset
	}

	proposals, err := dbmodels.FindApplicationApprovalRulesetBindingProposals(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding proposals", err, ginctx)
		return
	}

	proposal := dbmodels.CollectApplicationApprovalRulesetBindingVersionIDEquals(proposals, versionID)
	if proposal == nil {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": "application approval ruleset binding proposal not found"})
		return
	}

	otherProposals := dbmodels.CollectApplicationApprovalRulesetBindingVersionIDNotEquals(proposals, versionID)

	err = dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingVersionsPointerArray(proposals))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding adjustment", err, ginctx)
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
		json.PatchApplicationApprovalRulesetBindingAdjustment(orgID, &newAdjustment, input)

		if input.ProposalState == proposalstateinput.Final {
			proposalUpdate := proposal
			dbmodels.FinalizeReviewableProposal(&proposalUpdate.ReviewableVersionBase,
				&newAdjustment.ReviewableAdjustmentBase,
				latestApprovedVersionNumber,
				binding.CheckNewProposalsRequireReview(
					dbmodels.ReviewableActionUpdate,
					newAdjustment.Mode))
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
		creationRecord.ApplicationApprovalRulesetBindingVersionID = &proposal.ID
		creationRecord.ApplicationApprovalRulesetBindingAdjustmentNumber = &newAdjustment.AdjustmentNumber
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
				creationRecord.ApplicationApprovalRulesetBindingVersionID = &proposal.ID
				creationRecord.ApplicationApprovalRulesetBindingAdjustmentNumber = &newAdjustment.AdjustmentNumber
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

	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, proposal,
		appReadAuthorized, rulesetReadAuthorized)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateApplicationApprovalRulesetBindingProposalState(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ReviewableProposalStateInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appBindingReviewAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReviewApplicationApprovalRulesetBinding, application)
	appReadAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionReadApplication, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetBindingReviewAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReviewApprovalRulesetApplicationBinding, ruleset)
	rulesetReadAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset)

	if !appBindingReviewAuthorized || !rulesetBindingReviewAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	var latestApprovedVersionNumber uint32 = 0

	if input.State == reviewstateinput.Approved {
		dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.ApplicationApprovalRulesetBinding{&binding})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest approved version", err, ginctx)
			return
		}

		if binding.Version != nil {
			latestApprovedVersionNumber = *binding.Version.VersionNumber
		}
	}

	if appReadAuthorized {
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.Application{&application})
		if err != nil {
			respondWithDbQueryError("application latest version", err, ginctx)
			return
		}

		binding.Application = application
	}

	if rulesetReadAuthorized {
		err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
			[]*dbmodels.ApprovalRuleset{&ruleset})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest version", err, ginctx)
			return
		}

		binding.ApprovalRuleset = ruleset
	}

	proposals, err := dbmodels.FindApplicationApprovalRulesetBindingProposals(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding proposals", err, ginctx)
		return
	}

	proposal := dbmodels.CollectApplicationApprovalRulesetBindingVersionIDEquals(proposals, versionID)
	if proposal == nil {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": "application approval ruleset binding proposal not found"})
		return
	}

	otherProposals := dbmodels.CollectApplicationApprovalRulesetBindingVersionIDNotEquals(proposals, versionID)

	err = dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingVersionsPointerArray(proposals))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding adjustments", err, ginctx)
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
		creationRecord.ApplicationApprovalRulesetBindingVersionID = &proposal.ID
		creationRecord.ApplicationApprovalRulesetBindingAdjustmentNumber = &newAdjustment.AdjustmentNumber
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
				creationRecord.ApplicationApprovalRulesetBindingVersionID = &proposal.ID
				creationRecord.ApplicationApprovalRulesetBindingAdjustmentNumber = &newAdjustment.AdjustmentNumber
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

	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, proposal,
		appReadAuthorized, rulesetReadAuthorized)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) DeleteApplicationApprovalRulesetBindingProposal(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")
	rulesetID := ginctx.Param("ruleset_id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	appAuthorizer := authz.ApplicationAuthorizer{}
	appAuthorized := authz.AuthorizeSingularAction(appAuthorizer, orgMember, authz.ActionProposeBindApplicationToApprovalRuleset, application)
	rulesetAuthorizer := authz.ApprovalRulesetAuthorizer{}
	rulesetAuthorized := authz.AuthorizeSingularAction(rulesetAuthorizer, orgMember, authz.ActionProposeBindApprovalRulesetToApplication, ruleset)

	if !appAuthorized || !rulesetAuthorized {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(ctx.Db, orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	version, err := dbmodels.FindApplicationApprovalRulesetBindingProposalByID(ctx.Db, orgID, applicationID, rulesetID, versionID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding proposal", err, ginctx)
		return
	}
	binding.Version = &version

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		err = dbmodels.DeleteAuditCreationRecordsForApplicationApprovalRulesetBindingProposal(tx, orgID, version.ID)
		if err != nil {
			return err
		}
		err = dbmodels.DeleteApplicationApprovalRulesetBindingAdjustmentsForProposal(tx, orgID, version.ID)
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

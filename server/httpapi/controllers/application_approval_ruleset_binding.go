package controllers

import (
	"net/http"
	"strconv"

	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
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

	var input json.ApplicationApprovalRulesetBindingInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
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
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionManipulateApprovalRulesetBinding, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Modify database

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, rulesetID)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
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
	if input.Version.ProposalState == proposalstate.Final {
		dbmodels.FinalizeReviewableProposal(&version.ReviewableVersionBase,
			&adjustment.ReviewableAdjustmentBase, 0,
			binding.CheckNewProposalsRequireReview(
				dbmodels.ReviewableActionCreate,
				approvalrulesetbindingmode.Permissive))
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

	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, version, false, true)
	ginctx.JSON(http.StatusCreated, output)
}

func (ctx Context) ListApplicationApprovalRulesetBindings(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	applicationID := ginctx.Param("application_id")

	application, err := dbmodels.FindApplication(ctx.Db, orgID, applicationID)
	if err != nil {
		respondWithDbQueryError("application", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	tx, err := dbutils.ApplyDbQueryPagination(ginctx, ctx.Db)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	bindings, err := dbmodels.FindAllApplicationApprovalRulesetBindings(
		tx.Preload("ApprovalRuleset"),
		orgID, applicationID)
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
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApplicationApprovalRulesetBindingWithLatestApprovedVersion, 0, len(bindings))
	for _, binding := range bindings {
		outputList = append(outputList,
			json.CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(binding, binding.Version, false, true))
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

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, applicationID, rulesetID)
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

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApprovalRuleset{&binding.ApprovalRuleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest version", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateApplicationApprovalRulesetBindingWithLatestApprovedVersionAndAssociations(binding, binding.Version, false, true)
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

	var input json.ApplicationApprovalRulesetBindingInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionManipulateApprovalRulesetBinding, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, applicationID, rulesetID)
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

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApprovalRuleset{&binding.ApprovalRuleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest version", err, ginctx)
		return
	}

	var latestApprovedVersionNumber uint32 = 0

	if binding.Version != nil {
		latestApprovedVersionNumber = *binding.Version.VersionNumber
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		var oldMode = binding.Version.Adjustment.Mode
		var bindingUpdate dbmodels.ApplicationApprovalRulesetBinding = binding
		json.PatchApplicationApprovalRulesetBinding(&bindingUpdate, input)
		savetx := tx.Omit(clause.Associations).Model(&binding).Updates(bindingUpdate)
		if savetx.Error != nil {
			return savetx.Error
		}

		if input.Version != nil {
			newVersion, newAdjustment := binding.NewDraftVersion()
			if input.Version.ProposalState == proposalstate.Final {
				dbmodels.FinalizeReviewableProposal(&newVersion.ReviewableVersionBase,
					&newAdjustment.ReviewableAdjustmentBase,
					latestApprovedVersionNumber,
					binding.CheckNewProposalsRequireReview(dbmodels.ReviewableActionUpdate, oldMode))
			} else {
				dbmodels.SetReviewableAdjustmentReviewStateFromUnfinalizedProposalState(&newAdjustment.ReviewableAdjustmentBase,
					input.Version.ProposalState)
			}

			if err = tx.Omit(clause.Associations).Create(newVersion).Error; err != nil {
				return err
			}

			newAdjustment.ApplicationApprovalRulesetBindingVersionID = newVersion.ID
			json.PatchApplicationApprovalRulesetBindingAdjustment(orgID, newAdjustment, *input.Version)
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
		false, true)
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

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, application) {
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

	// Check authorization

	authorizer := authz.ApplicationAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApplication, application) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	binding, err := dbmodels.FindApplicationApprovalRulesetBinding(
		ctx.Db.Preload("ApprovalRuleset"),
		orgID, applicationID, rulesetID)
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&binding.ApprovalRuleset})
	if err != nil {
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
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
	output := json.CreateApplicationApprovalRulesetBindingWithVersionAndAssociations(binding, binding.Version, false, true)
	ginctx.JSON(http.StatusOK, output)
}

package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/authz"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json"
	"github.com/fullstaq-labs/sqedule/server/httpapi/json/proposalstate"
	reviewstateinput "github.com/fullstaq-labs/sqedule/server/httpapi/json/reviewstate"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//
// ******** Operations on resources ********
//

func (ctx Context) CreateApprovalRuleset(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	var input json.ApprovalRulesetInput
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

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionCreateApprovalRuleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Modify database

	ruleset := dbmodels.ApprovalRuleset{BaseModel: dbmodels.BaseModel{OrganizationID: orgID}}
	version, adjustment := ruleset.NewDraftVersion()
	adjustment.Rules = input.Version.ToDbmodelsApprovalRulesetContents(orgID)

	json.PatchApprovalRuleset(&ruleset, input)
	json.PatchApprovalRulesetAdjustment(orgID, adjustment, *input.Version)
	if input.Version.ProposalState == proposalstate.Final {
		dbmodels.FinalizeReviewableProposal(&version.ReviewableVersionBase,
			&adjustment.ReviewableAdjustmentBase, 0, false)
	}

	err := ctx.Db.Transaction(func(tx *gorm.DB) error {
		err := tx.Omit(clause.Associations).Create(&ruleset).Error
		if err != nil {
			return err
		}

		version.ApprovalRulesetID = ruleset.ID
		err = tx.Omit(clause.Associations).Create(version).Error
		if err != nil {
			return err
		}

		adjustment.ApprovalRulesetVersionID = version.ID
		err = tx.Omit(clause.Associations).Create(adjustment).Error
		if err != nil {
			return err
		}

		err = adjustment.Rules.ForEach(func(rule dbmodels.IApprovalRule) error {
			rule.AssociateWithApprovalRulesetAdjustment(*adjustment)
			return tx.Create(rule).Error
		})
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

	output := json.CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset, *version,
		*adjustment, []dbmodels.ApplicationApprovalRulesetBinding{}, []dbmodels.ReleaseApprovalRulesetBinding{},
		adjustment.Rules)
	ginctx.JSON(http.StatusCreated, output)
}

func (ctx Context) GetApprovalRulesets(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeCollectionAction(authorizer, orgMember, authz.ActionListApprovalRulesets) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	pagination, err := dbutils.ParsePaginationOptions(ginctx)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rulesets, err := dbmodels.FindAllApprovalRulesetsWithStats(ctx.Db, orgID, pagination)
	if err != nil {
		respondWithDbQueryError("approval rulesets", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithoutStats(rulesets))
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApprovalRulesetWithLatestApprovedVersion, 0, len(rulesets))
	for _, ruleset := range rulesets {
		outputList = append(outputList, json.CreateApprovalRulesetWithLatestApprovedVersionAndStats(ruleset,
			*ruleset.Version, *ruleset.Version.Adjustment))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApprovalRuleset(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID,
		[]*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsApprovalRules(ctx.Db, orgID,
		[]*dbmodels.ApprovalRulesetAdjustment{ruleset.Version.Adjustment})
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	appBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(appBindings))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	releaseBindings, err := dbmodels.FindAllReleaseApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Release.Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("release approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithReleases(dbmodels.CollectReleasesWithReleaseApprovalRulesetBindings(releaseBindings)))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateApprovalRulesetWithLatestApprovedVersionAndBindingsAndRules(ruleset, *ruleset.Version,
		*ruleset.Version.Adjustment, appBindings, releaseBindings, ruleset.Version.Adjustment.Rules)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateApprovalRuleset(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ApprovalRulesetInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	err = dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsApprovalRules(ctx.Db, orgID,
		[]*dbmodels.ApprovalRulesetAdjustment{ruleset.Version.Adjustment})
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	appBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(appBindings))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	releaseBindings, err := dbmodels.FindAllReleaseApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Release.Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("release approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithReleases(dbmodels.CollectReleasesWithReleaseApprovalRulesetBindings(releaseBindings)))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		var ruleset2 dbmodels.ApprovalRuleset = ruleset
		json.PatchApprovalRuleset(&ruleset2, input)
		savetx := tx.Omit(clause.Associations).Model(&ruleset).Updates(ruleset2)
		if savetx.Error != nil {
			return savetx.Error
		}

		if input.Version != nil {
			newVersion, newAdjustment := ruleset.NewDraftVersion()
			if input.Version.ProposalState == proposalstate.Final {
				dbmodels.FinalizeReviewableProposal(&newVersion.ReviewableVersionBase,
					&newAdjustment.ReviewableAdjustmentBase,
					*ruleset.Version.VersionNumber,
					ruleset.CheckNewProposalsRequireReview(len(appBindings) > 0))
			} else {
				dbmodels.SetReviewableAdjustmentReviewStateFromUnfinalizedProposalState(&newAdjustment.ReviewableAdjustmentBase,
					input.Version.ProposalState)
			}

			if err = tx.Omit(clause.Associations).Create(newVersion).Error; err != nil {
				return err
			}

			newAdjustment.ApprovalRulesetVersionID = newVersion.ID
			json.PatchApprovalRulesetAdjustment(orgID, newAdjustment, *input.Version)
			if err = tx.Omit(clause.Associations).Create(newAdjustment).Error; err != nil {
				return err
			}

			err = newAdjustment.Rules.ForEach(func(rule dbmodels.IApprovalRule) error {
				return tx.Omit(clause.Associations).Create(rule).Error
			})
			if err != nil {
				return err
			}

			ruleset.Version = newVersion
			ruleset.Version.Adjustment = newAdjustment
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset, *ruleset.Version,
		*ruleset.Version.Adjustment, appBindings, releaseBindings, ruleset.Version.Adjustment.Rules)
	ginctx.JSON(http.StatusOK, output)
}

//
// ******** Operations on approved versions ********
//

func (ctx Context) GetApprovalRulesetVersions(ginctx *gin.Context) {
	ctx.getApprovalRulesetVersionsOrProposals(ginctx, true)
}

func (ctx Context) getApprovalRulesetVersionsOrProposals(ginctx *gin.Context, approved bool) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	pagination, err := dbutils.ParsePaginationOptions(ginctx)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	versions, err := dbmodels.FindApprovalRulesetVersions(ctx.Db, orgID, id, approved, pagination)
	if err != nil {
		respondWithDbQueryError("approval ruleset versions", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, orgID, dbmodels.MakeApprovalRulesetVersionsPointerArray(versions))
	if err != nil {
		respondWithDbQueryError("approval ruleset adjustments", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsApprovalRules(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetAdjustmentsFromVersions(
			dbmodels.MakeApprovalRulesetVersionsPointerArray(versions)))
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsStats(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetAdjustmentsFromVersions(
			dbmodels.MakeApprovalRulesetVersionsPointerArray(versions)))
	if err != nil {
		respondWithDbQueryError("approval ruleset adjustment statistics", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApprovalRulesetVersion, 0, len(versions))
	for _, version := range versions {
		outputList = append(outputList, json.CreateApprovalRulesetVersionWithStatsAndRules(version, *version.Adjustment))
	}
	ginctx.JSON(http.StatusOK, gin.H{"items": outputList})
}

func (ctx Context) GetApprovalRulesetVersion(ginctx *gin.Context) {
	ctx.getApprovalRulesetVersionOrProposal(ginctx, true)
}

func (ctx Context) getApprovalRulesetVersionOrProposal(ginctx *gin.Context, approved bool) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")

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

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReadApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	var version dbmodels.ApprovalRulesetVersion
	if approved {
		version, err = dbmodels.FindApprovalRulesetVersionByNumber(ctx.Db, orgID, ruleset.ID, uint32(versionNumberOrID))
		if err != nil {
			respondWithDbQueryError("approval ruleset version", err, ginctx)
			return
		}
	} else {
		version, err = dbmodels.FindApprovalRulesetProposalByID(ctx.Db, orgID, ruleset.ID, versionNumberOrID)
		if err != nil {
			respondWithDbQueryError("approval ruleset proposal", err, ginctx)
			return
		}
	}
	ruleset.Version = &version

	err = dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, orgID, []*dbmodels.ApprovalRulesetVersion{&version})
	if err != nil {
		respondWithDbQueryError("approval ruleset adjustment", err, ginctx)
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsApprovalRules(ctx.Db, orgID,
		[]*dbmodels.ApprovalRulesetAdjustment{ruleset.Version.Adjustment})
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	appBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(appBindings))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	var releaseBindings []dbmodels.ReleaseApprovalRulesetBinding
	if approved {
		releaseBindings, err = dbmodels.FindAllReleaseApprovalRulesetBindingsWithApprovalRuleset(
			ctx.Db.Preload("Release.Application"), orgID, id)
		if err != nil {
			respondWithDbQueryError("release approval ruleset bindings", err, ginctx)
			return
		}
		err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
			dbmodels.CollectApplicationsWithReleases(dbmodels.CollectReleasesWithReleaseApprovalRulesetBindings(releaseBindings)))
		if err != nil {
			respondWithDbQueryError("application latest versions", err, ginctx)
			return
		}
	}

	// Generate response

	output := json.CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset, *ruleset.Version,
		*ruleset.Version.Adjustment, appBindings, releaseBindings, ruleset.Version.Adjustment.Rules)
	ginctx.JSON(http.StatusOK, output)
}

//
// ******** Operations on proposals ********
//

func (ctx Context) GetApprovalRulesetProposals(ginctx *gin.Context) {
	ctx.getApprovalRulesetVersionsOrProposals(ginctx, false)
}

func (ctx Context) GetApprovalRulesetProposal(ginctx *gin.Context) {
	ctx.getApprovalRulesetVersionOrProposal(ginctx, false)
}

func (ctx Context) UpdateApprovalRulesetProposal(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ApprovalRulesetVersionInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	var latestApprovedVersion dbmodels.IReviewableVersion
	dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest approved version", err, ginctx)
		return
	}
	latestApprovedVersion = ruleset.Version

	proposals, err := dbmodels.FindApprovalRulesetProposals(ctx.Db, orgID, ruleset.ID)
	if err != nil {
		respondWithDbQueryError("approval ruleset proposals", err, ginctx)
		return
	}

	proposal := dbmodels.CollectApprovalRulesetVersionIDEquals(proposals, versionID)
	if proposal == nil {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": "approval ruleset proposal not found"})
		return
	}
	ruleset.Version = proposal

	otherProposals := dbmodels.CollectApprovalRulesetVersionIDNotEquals(proposals, versionID)

	err = dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApprovalRulesetVersionsPointerArray(proposals))
	if err != nil {
		respondWithDbQueryError("approval ruleset adjustment", err, ginctx)
		return
	}

	if proposal.Adjustment.ReviewState == reviewstate.Reviewing && input.ProposalState == proposalstate.Final {
		ginctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Cannot finalize a proposal which is already being reviewed"})
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsApprovalRules(ctx.Db, orgID,
		[]*dbmodels.ApprovalRulesetAdjustment{ruleset.Version.Adjustment})
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	appBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(appBindings))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		// Create new Adjustment with patched state

		newAdjustment := proposal.Adjustment.NewAdjustment()

		if input.ProposalState == proposalstate.Final {
			proposalUpdate := proposal
			dbmodels.FinalizeReviewableProposal(&proposalUpdate.ReviewableVersionBase,
				&newAdjustment.ReviewableAdjustmentBase,
				*latestApprovedVersion.GetVersionNumber(),
				ruleset.CheckNewProposalsRequireReview(len(appBindings) > 0))
			if err = tx.Omit(clause.Associations).Model(&proposal).Updates(proposalUpdate).Error; err != nil {
				return err
			}
		} else {
			dbmodels.SetReviewableAdjustmentReviewStateFromUnfinalizedProposalState(&newAdjustment.ReviewableAdjustmentBase,
				input.ProposalState)
		}

		json.PatchApprovalRulesetAdjustment(orgID, &newAdjustment, input)
		if err = tx.Omit(clause.Associations).Create(&newAdjustment).Error; err != nil {
			return err
		}

		err = newAdjustment.Rules.ForEach(func(rule dbmodels.IApprovalRule) error {
			return tx.Omit(clause.Associations).Create(rule).Error
		})
		if err != nil {
			return err
		}

		proposal.Adjustment = &newAdjustment

		if newAdjustment.ReviewState == reviewstate.Approved {
			// Ensure no other proposals are in the Reviewing state

			for _, proposal := range otherProposals {
				if proposal.Adjustment.ReviewState != reviewstate.Reviewing {
					continue
				}

				newAdjustment := proposal.Adjustment.NewAdjustment()
				newAdjustment.ReviewState = reviewstate.Draft
				if err = tx.Omit(clause.Associations).Create(&newAdjustment).Error; err != nil {
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

	output := json.CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset, *ruleset.Version,
		*ruleset.Version.Adjustment, appBindings, []dbmodels.ReleaseApprovalRulesetBinding{}, ruleset.Version.Adjustment.Rules)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) UpdateApprovalRulesetProposalReviewState(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	var input json.ReviewableReviewStateInput
	if err := ginctx.ShouldBindJSON(&input); err != nil {
		ginctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionReviewApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	var latestApprovedVersion dbmodels.IReviewableVersion
	if input.State == reviewstateinput.Approved {
		dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&ruleset})
		if err != nil {
			respondWithDbQueryError("approval ruleset latest approved version", err, ginctx)
			return
		}
		latestApprovedVersion = ruleset.Version
	}

	proposals, err := dbmodels.FindApprovalRulesetProposals(ctx.Db, orgID, ruleset.ID)
	if err != nil {
		respondWithDbQueryError("approval ruleset proposals", err, ginctx)
		return
	}

	proposal := dbmodels.CollectApprovalRulesetVersionIDEquals(proposals, versionID)
	if proposal == nil {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": "approval ruleset proposal not found"})
		return
	}
	ruleset.Version = proposal

	otherProposals := dbmodels.CollectApprovalRulesetVersionIDNotEquals(proposals, versionID)

	err = dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, orgID,
		dbmodels.MakeApprovalRulesetVersionsPointerArray(proposals))
	if err != nil {
		respondWithDbQueryError("approval ruleset adjustments", err, ginctx)
		return
	}

	if proposal.Adjustment.ReviewState != reviewstate.Reviewing {
		ginctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": "This proposal is not awaiting review"})
		return
	}

	err = dbmodels.LoadApprovalRulesetAdjustmentsApprovalRules(ctx.Db, orgID,
		[]*dbmodels.ApprovalRulesetAdjustment{ruleset.Version.Adjustment})
	if err != nil {
		respondWithDbQueryError("approval rules", err, ginctx)
		return
	}

	appBindings, err := dbmodels.FindAllApplicationApprovalRulesetBindingsWithApprovalRuleset(
		ctx.Db.Preload("Application"), orgID, id)
	if err != nil {
		respondWithDbQueryError("application approval ruleset bindings", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithApplicationApprovalRulesetBindings(appBindings))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		// Create new Adjustment with new review state

		proposalUpdate := *proposal
		newAdjustment := proposal.Adjustment.NewAdjustment()

		if input.State == reviewstateinput.Approved {
			newAdjustment.ReviewState = reviewstate.Approved
			proposalUpdate.ApprovedAt = sql.NullTime{Time: time.Now(), Valid: true}
			proposalUpdate.VersionNumber = lib.NewUint32Ptr(*(latestApprovedVersion.GetVersionNumber()) + 1)
		} else {
			newAdjustment.ReviewState = reviewstate.Rejected
		}

		if err = tx.Omit(clause.Associations).Model(proposal).Updates(proposalUpdate).Error; err != nil {
			return err
		}
		if err = tx.Omit(clause.Associations).Create(&newAdjustment).Error; err != nil {
			return err
		}

		newAdjustment.Rules = proposal.Adjustment.Rules.CopyWithoutSavingAndAssociateWithAdjustment(newAdjustment)
		err = newAdjustment.Rules.ForEach(func(rule dbmodels.IApprovalRule) error {
			return tx.Omit(clause.Associations).Create(rule).Error
		})
		if err != nil {
			return err
		}

		proposal.Adjustment = &newAdjustment

		if input.State == reviewstateinput.Approved {
			// Ensure no other proposals are in the Reviewing state

			for _, proposal := range otherProposals {
				if proposal.Adjustment.ReviewState != reviewstate.Reviewing {
					continue
				}

				newAdjustment := proposal.Adjustment.NewAdjustment()
				newAdjustment.ReviewState = reviewstate.Draft
				if err = tx.Omit(clause.Associations).Create(&newAdjustment).Error; err != nil {
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

	output := json.CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset, *ruleset.Version,
		*ruleset.Version.Adjustment, appBindings, []dbmodels.ReleaseApprovalRulesetBinding{}, ruleset.Version.Adjustment.Rules)
	ginctx.JSON(http.StatusOK, output)
}

func (ctx Context) DeleteApprovalRulesetProposal(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")
	versionID, err := strconv.ParseUint(ginctx.Param("version_id"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_id' parameter as an integer: " + err.Error()})
		return
	}

	ruleset, err := dbmodels.FindApprovalRuleset(ctx.Db, orgID, id)
	if err != nil {
		respondWithDbQueryError("approval ruleset", err, ginctx)
		return
	}

	// Check authorization

	authorizer := authz.ApprovalRulesetAuthorizer{}
	if !authz.AuthorizeSingularAction(authorizer, orgMember, authz.ActionUpdateApprovalRuleset, ruleset) {
		respondWithUnauthorizedError(ginctx)
		return
	}

	// Query database

	version, err := dbmodels.FindApprovalRulesetProposalByID(ctx.Db, orgID, ruleset.ID, versionID)
	if err != nil {
		respondWithDbQueryError("approval ruleset proposal", err, ginctx)
		return
	}
	ruleset.Version = &version

	adjustments, err := dbmodels.FindApprovalRulesetAdjustments(ctx.Db, orgID, version.ID)
	if err != nil {
		respondWithDbQueryError("approval ruleset adjustments", err, ginctx)
		return
	}

	// Modify database

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		for i := range adjustments {
			err = ctx.Db.Delete(&adjustments[i]).Error
			if err != nil {
				return err
			}
		}

		err = ctx.Db.Delete(&version).Error
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

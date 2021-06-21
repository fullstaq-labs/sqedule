package controllers

import (
	"net/http"
	"strconv"

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
			version, adjustment := ruleset.NewDraftVersion()
			if input.Version.ProposalState == proposalstate.Final {
				dbmodels.FinalizeReviewableProposal(&version.ReviewableVersionBase,
					&adjustment.ReviewableAdjustmentBase,
					*ruleset.Version.VersionNumber,
					ruleset.CheckNewProposalsRequireReview(len(appBindings) > 0))
			}

			if err = tx.Omit(clause.Associations).Create(version).Error; err != nil {
				return err
			}

			adjustment.ApprovalRulesetVersionID = version.ID
			json.PatchApprovalRulesetAdjustment(orgID, adjustment, *input.Version)
			if err = tx.Omit(clause.Associations).Create(adjustment).Error; err != nil {
				return err
			}

			err = adjustment.Rules.ForEach(func(rule dbmodels.IApprovalRule) error {
				return tx.Omit(clause.Associations).Create(rule).Error
			})
			if err != nil {
				return err
			}

			ruleset.Version = version
			ruleset.Version.Adjustment = adjustment
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

func (ctx Context) GetApprovalRulesetVersion(ginctx *gin.Context) {
	// Fetch authentication, parse input, fetch related objects

	orgMember := auth.GetAuthenticatedOrgMemberNoFail(ginctx)
	orgID := orgMember.GetOrganizationID()
	id := ginctx.Param("id")

	versionNumber, err := strconv.ParseUint(ginctx.Param("version_number"), 10, 32)
	if err != nil {
		ginctx.JSON(http.StatusBadRequest,
			gin.H{"error": "Error parsing 'version_number' parameter as an integer: " + err.Error()})
		return
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

	version, err := dbmodels.FindApprovalRulesetVersionByNumber(ctx.Db, orgID, ruleset.ID, uint32(versionNumber))
	if err != nil {
		respondWithDbQueryError("approval ruleset version", err, ginctx)
		return
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

//
// ******** Operations on proposals ********
//

func (ctx Context) GetApprovalRulesetProposals(ginctx *gin.Context) {
	ctx.getApprovalRulesetVersionsOrProposals(ginctx, false)
}

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

	rulesetContents := input.Version.ToDbmodelsApprovalRulesetContents(orgID)
	ruleset := dbmodels.ApprovalRuleset{BaseModel: dbmodels.BaseModel{OrganizationID: orgID}}
	version, adjustment := ruleset.NewDraftVersion()

	json.PatchApprovalRuleset(&ruleset, input)
	json.PatchApprovalRulesetAdjustment(orgID, adjustment, &rulesetContents, *input.Version)
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

		err = rulesetContents.ForEach(func(rule dbmodels.IApprovalRule) error {
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
		rulesetContents)
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

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApprovalRulesetsWithoutStats(rulesets))
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	// Generate response

	outputList := make([]json.ApprovalRulesetWithLatestApprovedVersion, 0, len(rulesets))
	for _, ruleset := range rulesets {
		outputList = append(outputList, json.CreateApprovalRulesetWithLatestApprovedVersionAndStats(ruleset,
			*ruleset.LatestVersion, *ruleset.LatestAdjustment))
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

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID,
		[]*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	rules, err := dbmodels.FindApprovalRulesInRulesetVersion(ctx.Db, orgID, dbmodels.ApprovalRulesetVersionAndAdjustmentKey{
		VersionID:        ruleset.LatestVersion.ID,
		AdjustmentNumber: ruleset.LatestAdjustment.AdjustmentNumber,
	})
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
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
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
	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
		dbmodels.CollectApplicationsWithReleases(dbmodels.CollectReleasesWithReleaseApprovalRulesetBindings(releaseBindings)))
	if err != nil {
		respondWithDbQueryError("application latest versions", err, ginctx)
		return
	}

	// Generate response

	output := json.CreateApprovalRulesetWithLatestApprovedVersionAndBindingsAndRules(ruleset, *ruleset.LatestVersion,
		*ruleset.LatestAdjustment, appBindings, releaseBindings, rules)
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

	err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, orgID, []*dbmodels.ApprovalRuleset{&ruleset})
	if err != nil {
		respondWithDbQueryError("approval ruleset latest versions", err, ginctx)
		return
	}

	rules, err := dbmodels.FindApprovalRulesInRulesetVersion(ctx.Db, orgID, dbmodels.ApprovalRulesetVersionAndAdjustmentKey{
		VersionID:        ruleset.LatestVersion.ID,
		AdjustmentNumber: ruleset.LatestAdjustment.AdjustmentNumber,
	})
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
	err = dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersions(ctx.Db, orgID,
		dbmodels.MakeApplicationApprovalRulesetBindingsPointerArray(appBindings))
	if err != nil {
		respondWithDbQueryError("application approval ruleset binding latest versions", err, ginctx)
		return
	}
	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
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
	err = dbmodels.LoadApplicationsLatestVersions(ctx.Db, orgID,
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
					*ruleset.LatestVersion.VersionNumber,
					ruleset.CheckNewProposalsRequireReview(len(appBindings) > 0))
			}

			if err = tx.Omit(clause.Associations).Create(version).Error; err != nil {
				return err
			}

			adjustment.ApprovalRulesetVersionID = version.ID
			json.PatchApprovalRulesetAdjustment(orgID, adjustment, &rules, *input.Version)
			if err = tx.Omit(clause.Associations).Create(adjustment).Error; err != nil {
				return err
			}

			err = rules.ForEach(func(rule dbmodels.IApprovalRule) error {
				return tx.Omit(clause.Associations).Create(rule).Error
			})
			if err != nil {
				return err
			}

			ruleset.LatestVersion = version
			ruleset.LatestAdjustment = adjustment
		}

		return nil
	})
	if err != nil {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate response

	output := json.CreateApprovalRulesetWithVersionAndBindingsAndRules(ruleset, *ruleset.LatestVersion,
		*ruleset.LatestAdjustment, appBindings, releaseBindings, rules)
	ginctx.JSON(http.StatusOK, output)
}

package controllers

import "github.com/gin-gonic/gin"

func (ctx Context) InstallRoutes(rg *gin.RouterGroup) {
	// Organizations
	rg.GET("organization", ctx.GetCurrentOrganization)
	rg.PATCH("organization", ctx.UpdateCurrentOrganization)
	rg.GET("organizations/:id", ctx.GetOrganization)
	rg.PATCH("organizations/:id", ctx.UpdateOrganization)

	// Applications
	rg.GET("applications", ctx.GetApplications)
	rg.GET("applications/:application_id", ctx.GetApplication)

	// Releases
	rg.GET("releases", ctx.GetReleases)
	rg.GET("applications/:application_id/releases", ctx.GetReleases)
	rg.POST("applications/:application_id/releases", ctx.CreateRelease)
	rg.GET("applications/:application_id/releases/:id", ctx.GetRelease)
	rg.PATCH("applications/:application_id/releases/:id", ctx.UpdateRelease)

	// Approval ruleset bindings
	rg.GET("applications/:application_id/approval-ruleset-bindings", ctx.GetApplicationApprovalRulesetBindings)

	// Approval rulesets
	rg.POST("approval-rulesets", ctx.CreateApprovalRuleset)
	rg.GET("approval-rulesets", ctx.GetApprovalRulesets)
	rg.GET("approval-rulesets/:id", ctx.GetApprovalRuleset)
	rg.PATCH("approval-rulesets/:id", ctx.UpdateApprovalRuleset)
	rg.GET("approval-rulesets/:id/versions", ctx.GetApprovalRulesetVersions)
	rg.GET("approval-rulesets/:id/versions/:version_number", ctx.GetApprovalRulesetVersion)
	rg.GET("approval-rulesets/:id/proposals", ctx.GetApprovalRulesetProposals)
	rg.GET("approval-rulesets/:id/proposals/:version_id", ctx.GetApprovalRulesetProposal)
	rg.PATCH("approval-rulesets/:id/proposals/:version_id", ctx.UpdateApprovalRulesetProposal)
	rg.PUT("approval-rulesets/:id/proposals/:version_id/review-state", ctx.UpdateApprovalRulesetProposalReviewState)
	rg.DELETE("approval-rulesets/:id/proposals/:version_id", ctx.DeleteApprovalRulesetProposal)
}

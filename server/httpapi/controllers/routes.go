package controllers

import "github.com/gin-gonic/gin"

func (ctx Context) InstallRoutes(rg *gin.RouterGroup) {
	// Organizations
	rg.GET("organization", ctx.GetCurrentOrganization)
	rg.PATCH("organization", ctx.UpdateCurrentOrganization)
	rg.GET("organizations/:id", ctx.GetOrganization)
	rg.PATCH("organizations/:id", ctx.UpdateOrganization)

	// Applications
	rg.GET("applications", ctx.ListApplications)
	rg.GET("applications/:application_id", ctx.GetApplication)

	// Releases
	rg.GET("releases", ctx.ListReleases)
	rg.GET("applications/:application_id/releases", ctx.ListReleases)
	rg.POST("applications/:application_id/releases", ctx.CreateRelease)
	rg.GET("applications/:application_id/releases/:id", ctx.GetRelease)
	rg.PATCH("applications/:application_id/releases/:id", ctx.UpdateRelease)

	// Approval ruleset bindings
	rg.POST("applications/:application_id/approval-ruleset-bindings/:ruleset_id", ctx.CreateApplicationApprovalRulesetBinding)
	rg.GET("applications/:application_id/approval-ruleset-bindings", ctx.ListApplicationApprovalRulesetBindings)
	rg.GET("applications/:application_id/approval-ruleset-bindings/:ruleset_id", ctx.GetApplicationApprovalRulesetBinding)
	rg.PATCH("applications/:application_id/approval-ruleset-bindings/:ruleset_id", ctx.UpdateApplicationApprovalRulesetBinding)
	rg.GET("applications/:application_id/approval-ruleset-bindings/:ruleset_id/versions", ctx.ListApplicationApprovalRulesetBindingVersions)
	rg.GET("applications/:application_id/approval-ruleset-bindings/:ruleset_id/versions/:version_number", ctx.GetApplicationApprovalRulesetBindingVersion)
	rg.GET("applications/:application_id/approval-ruleset-bindings/:ruleset_id/proposals", ctx.ListApplicationApprovalRulesetBindingProposals)
	rg.GET("applications/:application_id/approval-ruleset-bindings/:ruleset_id/proposals/:version_id", ctx.GetApplicationApprovalRulesetBindingProposal)

	// Approval rulesets
	rg.POST("approval-rulesets", ctx.CreateApprovalRuleset)
	rg.GET("approval-rulesets", ctx.ListApprovalRulesets)
	rg.GET("approval-rulesets/:id", ctx.GetApprovalRuleset)
	rg.PATCH("approval-rulesets/:id", ctx.UpdateApprovalRuleset)
	rg.GET("approval-rulesets/:id/versions", ctx.ListApprovalRulesetVersions)
	rg.GET("approval-rulesets/:id/versions/:version_number", ctx.GetApprovalRulesetVersion)
	rg.GET("approval-rulesets/:id/proposals", ctx.ListApprovalRulesetProposals)
	rg.GET("approval-rulesets/:id/proposals/:version_id", ctx.GetApprovalRulesetProposal)
	rg.PATCH("approval-rulesets/:id/proposals/:version_id", ctx.UpdateApprovalRulesetProposal)
	rg.PUT("approval-rulesets/:id/proposals/:version_id/review-state", ctx.UpdateApprovalRulesetProposalReviewState)
	rg.DELETE("approval-rulesets/:id/proposals/:version_id", ctx.DeleteApprovalRulesetProposal)
}

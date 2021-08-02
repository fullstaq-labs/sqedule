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
	rg.POST("applications", ctx.CreateApplication)
	rg.GET("applications/:application_id", ctx.GetApplication)
	rg.PATCH("applications/:application_id", ctx.UpdateApplication)
	rg.GET("applications/:application_id/versions", ctx.ListApplicationVersions)
	rg.GET("applications/:application_id/versions/:version_number", ctx.GetApplicationVersion)
	rg.GET("applications/:application_id/proposals", ctx.ListApplicationProposals)
	rg.GET("applications/:application_id/proposals/:version_id", ctx.GetApplicationProposal)
	rg.PATCH("applications/:application_id/proposals/:version_id", ctx.UpdateApplicationProposal)
	rg.PUT("applications/:application_id/proposals/:version_id/review-state", ctx.UpdateApplicationProposalReviewState)
	rg.DELETE("applications/:application_id/proposals/:version_id", ctx.DeleteApplicationProposal)

	// Releases
	rg.GET("releases", ctx.ListReleases)
	rg.GET("applications/:application_id/releases", ctx.ListReleases)
	rg.POST("applications/:application_id/releases", ctx.CreateRelease)
	rg.GET("applications/:application_id/releases/:id", ctx.GetRelease)
	rg.PATCH("applications/:application_id/releases/:id", ctx.UpdateRelease)

	// Approval ruleset bindings
	rg.GET("application-approval-ruleset-bindings", ctx.ListApplicationApprovalRulesetBindings)
	rg.POST("application-approval-ruleset-bindings/:application_id/:ruleset_id", ctx.CreateApplicationApprovalRulesetBinding)
	rg.GET("application-approval-ruleset-bindings/:application_id/:ruleset_id", ctx.GetApplicationApprovalRulesetBinding)
	rg.PATCH("application-approval-ruleset-bindings/:application_id/:ruleset_id", ctx.UpdateApplicationApprovalRulesetBinding)
	rg.GET("application-approval-ruleset-bindings/:application_id/:ruleset_id/versions", ctx.ListApplicationApprovalRulesetBindingVersions)
	rg.GET("application-approval-ruleset-bindings/:application_id/:ruleset_id/versions/:version_number", ctx.GetApplicationApprovalRulesetBindingVersion)
	rg.GET("application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals", ctx.ListApplicationApprovalRulesetBindingProposals)
	rg.GET("application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id", ctx.GetApplicationApprovalRulesetBindingProposal)
	rg.PATCH("application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id", ctx.UpdateApplicationApprovalRulesetBindingProposal)
	rg.PUT("application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id/review-state", ctx.UpdateApplicationApprovalRulesetBindingProposalReviewState)
	rg.DELETE("application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id", ctx.DeleteApplicationApprovalRulesetBindingProposal)
	rg.GET("applications/:application_id/approval-ruleset-bindings", ctx.ListApplicationApprovalRulesetBindings)

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

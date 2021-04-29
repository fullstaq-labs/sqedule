package controllers

import "github.com/gin-gonic/gin"

func (ctx Context) InstallRoutes(rg *gin.RouterGroup) {
	// Organizations
	rg.GET("organization", ctx.GetCurrentOrganization)
	rg.PATCH("organization", ctx.PatchCurrentOrganization)
	rg.GET("organizations/:id", ctx.GetOrganization)
	rg.PATCH("organizations/:id", ctx.PatchOrganization)

	// Applications
	rg.GET("applications", ctx.GetAllApplications)
	rg.GET("applications/:application_id", ctx.GetApplication)

	// Releases
	rg.GET("releases", ctx.GetAllReleases)
	rg.GET("applications/:application_id/releases", ctx.GetAllReleases)
	rg.POST("applications/:application_id/releases", ctx.CreateRelease)
	rg.GET("applications/:application_id/releases/:id", ctx.GetRelease)
	rg.PATCH("applications/:application_id/releases/:id", ctx.PatchRelease)

	// Approval rulesets
	rg.GET("applications/:application_id/approval-ruleset-bindings", ctx.GetAllApplicationApprovalRulesetBindings)
	rg.GET("approval-rulesets", ctx.GetAllApprovalRulesets)
	rg.GET("approval-rulesets/:id", ctx.GetApprovalRuleset)
}

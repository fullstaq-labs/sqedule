package httpapi

import (
	"fmt"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter ...
func (ctx Context) SetupRouter(engine *gin.Engine) error {
	jwtAuthMiddleware, err := ctx.createJwtAuthMiddleware()
	if err != nil {
		return fmt.Errorf("error setting up authentication middleware: %w", err)
	}

	engine.Use(cors.New(ctx.createCorsConfig()))

	v1 := engine.Group("/v1")
	authenticatedGroup := v1.Group("/")
	if !ctx.UseTestAuthentication {
		authenticatedGroup.Use(jwtAuthMiddleware.MiddlewareFunc())
	}
	authenticatedGroup.Use(ctx.lookupAndRequireAuthenticatedOrganizationMember)

	ctx.setupUnauthenticatedRoutes(v1, jwtAuthMiddleware)
	ctx.setupAuthenticatedRoutes(authenticatedGroup)
	return nil
}

func (ctx Context) setupUnauthenticatedRoutes(rg *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	rg.POST("/auth/login", authMiddleware.LoginHandler)
	rg.POST("/auth/refresh-token", authMiddleware.RefreshHandler)
}

func (ctx Context) setupAuthenticatedRoutes(rg *gin.RouterGroup) {
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

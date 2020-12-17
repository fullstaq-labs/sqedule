package httpapi

import (
	"fmt"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

// SetupRouter ...
func (ctx Context) SetupRouter(engine *gin.Engine) error {
	authMiddleware, err := ctx.createAuthMiddleware()
	if err != nil {
		return fmt.Errorf("error setting up authentication middleware: %w", err)
	}

	v1 := engine.Group("/v1")
	authenticatedGroup := v1.Group("/")
	authenticatedGroup.Use(authMiddleware.MiddlewareFunc())
	authenticatedGroup.Use(ctx.lookupAndRequireAuthenticatedOrganizationMember)

	ctx.setupUnauthenticatedRoutes(v1, authMiddleware)
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
	rg.GET("organizations/:name", ctx.GetOrganization)
	rg.PATCH("organizations/:name", ctx.PatchOrganization)
}

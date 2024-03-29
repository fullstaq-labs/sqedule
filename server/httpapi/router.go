package httpapi

import (
	"fmt"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/fullstaq-labs/sqedule/server/httpapi/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	gormlogger "gorm.io/gorm/logger"
)

func (ctx Context) SetupRouter(engine *gin.Engine, logger gormlogger.Interface) error {
	controllerCtx := controllers.NewContext(ctx.Db, ctx.WaitGroup)
	jwtAuthMiddleware, orgMemberLookupMiddleware, err := ctx.newAuthMiddlewares()
	if err != nil {
		return err
	}

	if corsConfig, ok := ctx.createCorsConfig(logger); ok {
		engine.Use(cors.New(corsConfig))
	}

	v1 := engine.Group("/v1")
	ctx.installUnauthenticatedRoutes(v1, jwtAuthMiddleware, controllerCtx)

	authenticatedGroup := v1.Group("/")
	ctx.installAuthenticationMiddlewares(authenticatedGroup, jwtAuthMiddleware, orgMemberLookupMiddleware)
	ctx.installAuthenticatedRoutes(authenticatedGroup, controllerCtx)
	return nil
}

func (ctx Context) newAuthMiddlewares() (*jwt.GinJWTMiddleware, gin.HandlerFunc, error) {
	jwtAuthMiddleware, err := auth.NewJwtMiddleware(ctx.Db)
	if err != nil {
		return nil, nil, fmt.Errorf("error setting up JWT authentication middleware: %w", err)
	}

	orgMemberLookupMiddleware := auth.NewOrgMemberLookupMiddleware(ctx.Db, ctx.UseTestAuthentication)

	return jwtAuthMiddleware, orgMemberLookupMiddleware, nil
}

func (ctx Context) installUnauthenticatedRoutes(rg *gin.RouterGroup, jwtAuthMiddleware *jwt.GinJWTMiddleware, controllerCtx controllers.Context) {
	rg.POST("/auth/login", jwtAuthMiddleware.LoginHandler)
	rg.POST("/auth/refresh-token", jwtAuthMiddleware.RefreshHandler)
	controllerCtx.InstallUnauthenticatedRoutes(rg)
}

func (ctx Context) installAuthenticationMiddlewares(rg *gin.RouterGroup, jwtAuthMiddleware *jwt.GinJWTMiddleware, orgMemberLookupMiddleware gin.HandlerFunc) {
	// TODO: uncomment this once we have proper user management
	// if !ctx.UseTestAuthentication {
	// 	rg.Use(jwtAuthMiddleware.MiddlewareFunc())
	// }
	rg.Use(orgMemberLookupMiddleware)
}

func (ctx Context) installAuthenticatedRoutes(rg *gin.RouterGroup, controllerCtx controllers.Context) {
	controllerCtx.InstallAuthenticatedRoutes(rg)
}

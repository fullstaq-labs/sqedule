package httpapi

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	gormlogger "gorm.io/gorm/logger"
)

func (ctx Context) createCorsConfig(logger gormlogger.Interface) (cors.Config, bool) {
	var allowedOrigins []string
	if ctx.DevelopmentMode {
		allowedOrigins = append(allowedOrigins, "http://localhost:3000")
		allowedOrigins = append(allowedOrigins, "https://localhost:3000")
		allowedOrigins = append(allowedOrigins, "http://127.0.0.1:3000")
		allowedOrigins = append(allowedOrigins, "https://127.0.0.1:3000")
		logger.Info(context.Background(), "Development mode enabled, so allowing CORS origins: http(s)://{localhost,127.0.0.1}:3000")
		logger.Info(context.Background(), "If the Next.js server is running on a different port, then please set --cors-origin")
	}
	if len(ctx.CorsOrigin) > 0 {
		allowedOrigins = append(allowedOrigins, ctx.CorsOrigin)
	}

	if len(allowedOrigins) > 0 {
		return cors.Config{
			AllowOrigins:  allowedOrigins,
			AllowMethods:  []string{"HEAD", "GET", "POST", "PATCH", "DELETE"},
			AllowHeaders:  []string{"Authorization", "Content-Length", "Content-Type"},
			ExposeHeaders: []string{"Content-Length"},
			MaxAge:        12 * time.Hour,
		}, true
	} else {
		return cors.Config{}, false
	}
}

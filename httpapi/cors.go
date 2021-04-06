package httpapi

import (
	"time"

	"github.com/gin-contrib/cors"
)

func (ctx Context) createCorsConfig() cors.Config {
	allowedOrigins := []string{"https://localhost:3000"}
	if len(ctx.CorsOrigin) > 0 {
		allowedOrigins = append(allowedOrigins, ctx.CorsOrigin)
	}

	return cors.Config{
		// TODO: change this
		AllowOrigins:  allowedOrigins,
		AllowMethods:  []string{"HEAD", "GET", "POST", "PATCH", "DELETE"},
		AllowHeaders:  []string{"Authorization", "Content-Length", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}
}

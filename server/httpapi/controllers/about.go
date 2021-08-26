package controllers

import (
	"net/http"

	"github.com/fullstaq-labs/sqedule/server"
	"github.com/gin-gonic/gin"
)

func (ctx Context) About(ginctx *gin.Context) {
	ginctx.JSON(http.StatusOK, gin.H{
		"version": server.VersionString,
	})
}

package controllers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func respondWithDbQueryError(resourceTypeName string, err error, ginctx *gin.Context) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ginctx.JSON(http.StatusNotFound, gin.H{"error": resourceTypeName + " not found"})
	} else {
		ginctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error querying database for " + resourceTypeName})
	}
}

func respondWithUnauthorizedError(ginctx *gin.Context) {
	ginctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized action"})
}

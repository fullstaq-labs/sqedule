package httpapi

import (
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const orgMemberContextKey = "authenticated_organization_member"

// GetAuthenticatedOrganizationMemberNoFail returns the OrganizationMember that's
// authenticated with the current request.
//
// It assumes that the lookupAndRequireAuthenticatedOrganizationMember middleware
// has already run, hence this function never returns an error.
func GetAuthenticatedOrganizationMemberNoFail(ginctx *gin.Context) dbmodels.IOrganizationMember {
	orgMember, exists := ginctx.Get(orgMemberContextKey)
	if !exists {
		panic("Bug: no authenticated organization member found in request context. Does the request use the lookupAndRequireAuthenticatedOrganizationMember() middleware?")
	}
	return orgMember.(dbmodels.IOrganizationMember)
}

// lookupAndRequireAuthenticatedOrganizationMember finds out which organization member
// is authenticated with the current requests, looks up its record from the database,
// and associates that record with the current request.
//
// You can get the looked up record using `GetAuthenticatedOrganizationMemberNoFail()`.
//
// If the OrganizationMember cannot be looked up, or if some other error occurs,
// then aborts the request with an error.
func (ctx Context) lookupAndRequireAuthenticatedOrganizationMember(ginctx *gin.Context) {
	orgMember, err := ctx.lookupTestAuthenticatedOrganizationMember(ginctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		if err != gorm.ErrRecordNotFound {
			ginctx.Abort()
			ginctx.JSON(http.StatusInternalServerError,
				gin.H{"error": fmt.Sprintf("internal authentication error: internal database error: %s", err.Error())})
			return
		}
	} else if orgMember != nil {
		ginctx.Set(orgMemberContextKey, orgMember)
		ginctx.Next()
		return
	}

	orgID, orgMemberType, orgMemberID, ok := getOrgMemberFromJwtClaims(ginctx)
	if !ok {
		ginctx.Abort()
		ginctx.JSON(http.StatusUnauthorized,
			gin.H{"error": "authentication error: incomplete JWT token"})
		return
	}

	orgMember, err = dbmodels.FindOrganizationMember(ctx.Db, orgID, orgMemberType, orgMemberID)
	if err != nil {
		ginctx.Abort()
		if err == gorm.ErrRecordNotFound {
			ginctx.JSON(http.StatusUnauthorized,
				gin.H{"error": "no database record found for authenticated organization member"})
		} else {
			ginctx.JSON(http.StatusInternalServerError,
				gin.H{"error": "internal authentication error: internal database error"})
		}
		return
	}

	ginctx.Set(orgMemberContextKey, orgMember)
	ginctx.Next()
}

func (ctx Context) lookupTestAuthenticatedOrganizationMember(ginctx *gin.Context) (dbmodels.IOrganizationMember, error) {
	if !ctx.UseTestAuthentication {
		return nil, nil
	}
	orgID := ginctx.GetHeader("TestOrgID")
	orgMemberType := ginctx.GetHeader("TestOrgMemberType")
	orgMemberID := ginctx.GetHeader("TestOrgMemberID")
	if len(orgID) == 0 || len(orgMemberType) == 0 || len(orgMemberID) == 0 {
		return nil, nil
	}
	return dbmodels.FindOrganizationMember(ctx.Db, orgID, dbmodels.OrganizationMemberType(orgMemberType), orgMemberID)
}

func getOrgMemberFromJwtClaims(ginctx *gin.Context) (string, dbmodels.OrganizationMemberType, string, bool) {
	var orgID, orgMemberType, orgMemberID string
	var ok bool
	claims := jwt.ExtractClaims(ginctx)

	orgID, ok = claims[orgIDClaim].(string)
	if ok {
		orgMemberType, ok = claims[orgMemberTypeClaim].(string)
	}
	if ok {
		orgMemberID, ok = claims[orgMemberIDClaim].(string)
	}
	if ok {
		return orgID, dbmodels.OrganizationMemberType(orgMemberType), orgMemberID, true
	}
	return "", "", "", false
}

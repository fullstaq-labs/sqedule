package auth

import (
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const OrgMemberContextKey = "authenticated_organization_member"

// NewOrgMemberLookupMiddleware returns a Gin middleware which finds out which organization member
// is authenticated with the current request, looks up its record from the database,
// and associates that record with the current request.
//
// If `testing` is true, then it determines which organization member is authenticated by
// looking at the HTTP headers `TestOrgID`, `TestOrgMemberType` and `TestOrgMemberID`.
//
// If `testing` is false, or if no HTTP headers specify which organization member to authenticate
// as, then it looks at the JWT authorization token. This requires that the `NewJwtMiddleware()`
// middleware has already run.
//
// You can get the looked up record using `GetAuthenticatedOrgMemberNoFail()`.
//
// If no OrganizationMember is found at the end, or if some other error occurs,
// then this middleware aborts the request with an error.
func NewOrgMemberLookupMiddleware(db *gorm.DB, testing bool) gin.HandlerFunc {
	m := orgMemberLookupMiddleware{Db: db, Testing: testing}
	return func(ginctx *gin.Context) {
		m.run(ginctx)
	}
}

// GetAuthenticatedOrgMemberNoFail returns the OrganizationMember that's
// authenticated with the current request.
//
// It assumes that the `NewOrgMemberLookupMiddleware()` middleware
// has already run, hence this function never returns an error.
func GetAuthenticatedOrgMemberNoFail(ginctx *gin.Context) dbmodels.IOrganizationMember {
	orgMember, exists := ginctx.Get(OrgMemberContextKey)
	if !exists {
		panic("Bug: no authenticated organization member found in request context. Does the request use the NewOrgMemberLookupMiddleware() middleware?")
	}
	return orgMember.(dbmodels.IOrganizationMember)
}

type orgMemberLookupMiddleware struct {
	Db      *gorm.DB
	Testing bool
}

func (m orgMemberLookupMiddleware) run(ginctx *gin.Context) {
	orgMember, err := m.lookupTestAuthenticatedOrgMember(ginctx)
	if err != nil && err != gorm.ErrRecordNotFound {
		if err != gorm.ErrRecordNotFound {
			ginctx.Abort()
			ginctx.JSON(http.StatusInternalServerError,
				gin.H{"error": fmt.Sprintf("internal authentication error: internal database error: %s", err.Error())})
			return
		}
	} else if orgMember != nil {
		ginctx.Set(OrgMemberContextKey, orgMember)
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

	orgMember, err = dbmodels.FindOrganizationMember(m.Db, orgID, orgMemberType, orgMemberID)
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

	ginctx.Set(OrgMemberContextKey, orgMember)
	ginctx.Next()
}

func (m orgMemberLookupMiddleware) lookupTestAuthenticatedOrgMember(ginctx *gin.Context) (dbmodels.IOrganizationMember, error) {
	if !m.Testing {
		return nil, nil
	}

	orgID := ginctx.GetHeader("TestOrgID")
	orgMemberType := ginctx.GetHeader("TestOrgMemberType")
	orgMemberID := ginctx.GetHeader("TestOrgMemberID")
	if len(orgID) == 0 || len(orgMemberType) == 0 || len(orgMemberID) == 0 {
		return nil, nil
	}

	return dbmodels.FindOrganizationMember(m.Db, orgID, dbmodels.OrganizationMemberType(orgMemberType), orgMemberID)
}

func getOrgMemberFromJwtClaims(ginctx *gin.Context) (string, dbmodels.OrganizationMemberType, string, bool) {
	var orgID, orgMemberType, orgMemberID string
	var ok bool
	claims := jwt.ExtractClaims(ginctx)

	orgID, ok = claims[jwtOrgIDClaim].(string)
	if ok {
		orgMemberType, ok = claims[jwtOrgMemberTypeClaim].(string)
	}
	if ok {
		orgMemberID, ok = claims[jwtOrgMemberIDClaim].(string)
	}
	if ok {
		return orgID, dbmodels.OrganizationMemberType(orgMemberType), orgMemberID, true
	}
	return "", "", "", false
}

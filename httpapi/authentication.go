package httpapi

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	orgIDClaim         = "orgid"
	orgMemberTypeClaim = "omt"
	orgMemberIDClaim   = "omid"

	orgMemberContextKey = "authenticated_organization_member"
)

func (ctx Context) createAuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:         "Sqedule",
		Key:           []byte("secret key"),
		Timeout:       time.Hour * 1024,
		MaxRefresh:    time.Hour * 1024,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		Authenticator: ctx.jwtLoginOrganizationMember,
		PayloadFunc:   jwtConvertOrgMemberToClaims,
	})
}

// lookupAndRequireAuthenticatedOrganizationMember finds out which organization member
// is authenticated with the current requests, looks up its record from the database,
// and associates that record with the current request.
//
// You can get the looked up record using `getGuaranteedAuthenticatedOrganizationMember()`.
//
// If the OrganizationMember cannot be looked up, or if some other error occurs,
// then aborts the request with an error.
func (ctx Context) lookupAndRequireAuthenticatedOrganizationMember(ginctx *gin.Context) {
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
	if !ok {
		ginctx.Abort()
		ginctx.JSON(http.StatusUnauthorized,
			gin.H{"error": "authentication error: incomplete JWT token"})
		return
	}

	var orgMember dbmodels.IOrganizationMember
	var err error

	switch orgMemberType {
	case dbmodels.UserTypeShortName:
		orgMember, err = dbmodels.FindUserByEmail(ctx.Db, orgID, orgMemberID)
	case dbmodels.ServiceAccountTypeShortName:
		orgMember, err = dbmodels.FindServiceAccountByName(ctx.Db, orgID, orgMemberID)
	default:
		panic(fmt.Errorf("Bug: unsupported organization member type %s", orgMemberType))
	}

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

// getAuthenticatedOrganizationMemberNoFail returns the OrganizationMember that's
// authenticated with the current request. It assumes that the
// lookupAndRequireAuthenticatedOrganizationMember middleware has already run,
// hence this function never returns an error.
func getAuthenticatedOrganizationMemberNoFail(ginctx *gin.Context) dbmodels.IOrganizationMember {
	orgMember, exists := ginctx.Get(orgMemberContextKey)
	if !exists {
		panic("Bug: no authenticated organization member found in request context. Does the request use the lookupAndRequireAuthenticatedOrganizationMember() middleware?")
	}
	return orgMember.(dbmodels.IOrganizationMember)
}

type loginVals struct {
	OrganizationID     string `json:"organization_id"`
	Email              string `json:"email"`
	ServiceAccountName string `json:"service_account_name"`
	AccessToken        string `json:"access_token"`
}

func (ctx Context) jwtLoginOrganizationMember(ginctx *gin.Context) (interface{}, error) {
	var loginVals loginVals
	var orgMember dbmodels.IOrganizationMember
	var err error

	if err = ginctx.ShouldBind(&loginVals); err != nil {
		return nil, jwt.ErrMissingLoginValues
	}
	if err = jwtValidateLoginVals(loginVals); err != nil {
		return nil, err
	}

	if orgMember, err = ctx.jwtLookupOrganizationMemberWithLoginVals(loginVals); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("incorrect organization ID or %s", orgMember.IDTypeDisplayName())
		}
		return nil, errors.New("internal database error")
	}

	ok, err := orgMember.Authenticate(loginVals.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("error authenticating user: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("incorrect access token")
	}

	return orgMember, nil
}

func (ctx Context) jwtLookupOrganizationMemberWithLoginVals(loginVals loginVals) (dbmodels.IOrganizationMember, error) {
	if len(loginVals.Email) > 0 {
		return dbmodels.FindUserByEmail(ctx.Db, loginVals.OrganizationID,
			loginVals.Email)
	} else if len(loginVals.ServiceAccountName) > 0 {
		return dbmodels.FindServiceAccountByName(ctx.Db, loginVals.OrganizationID,
			loginVals.ServiceAccountName)
	} else {
		panic("Bug")
	}
}

func jwtValidateLoginVals(loginVals loginVals) error {
	if len(loginVals.OrganizationID) == 0 {
		return errors.New("missing organization ID")
	}
	if len(loginVals.Email) == 0 && len(loginVals.ServiceAccountName) == 0 {
		return errors.New("missing login ID")
	}
	if len(loginVals.Email) > 0 && len(loginVals.ServiceAccountName) > 0 {
		return errors.New("only email or service_account_name may be specified, not both")
	}
	if len(loginVals.AccessToken) == 0 {
		return errors.New("missing access token")
	}
	return nil
}

func jwtConvertOrgMemberToClaims(data interface{}) jwt.MapClaims {
	orgMember := data.(dbmodels.IOrganizationMember)
	return jwt.MapClaims{
		orgIDClaim:         orgMember.GetOrganizationMember().BaseModel.OrganizationID,
		orgMemberTypeClaim: orgMember.TypeShortName(),
		orgMemberIDClaim:   orgMember.ID(),
	}
}

package httpapi

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	orgIDClaim         = "orgid"
	orgMemberTypeClaim = "omt"
	orgMemberIDClaim   = "omid"
)

func (ctx Context) createJwtAuthMiddleware() (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:         "Sqedule",
		Key:           []byte("secret key"),
		Timeout:       time.Hour * 24 * 365,
		MaxRefresh:    time.Hour * 24 * 365,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		Authenticator: ctx.jwtLoginOrganizationMember,
		PayloadFunc:   jwtConvertOrgMemberToClaims,
	})
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
		orgMemberTypeClaim: orgMember.Type(),
		orgMemberIDClaim:   orgMember.ID(),
	}
}

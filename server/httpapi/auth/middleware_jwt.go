package auth

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
	jwtOrgIDClaim         = "orgid"
	jwtOrgMemberTypeClaim = "omt"
	jwtOrgMemberIDClaim   = "omid"
)

func NewJwtMiddleware(db *gorm.DB) (*jwt.GinJWTMiddleware, error) {
	m := jwtMiddleware{Db: db}
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:         "Sqedule",
		Key:           []byte("secret key"),
		Timeout:       time.Hour * 24 * 365,
		MaxRefresh:    time.Hour * 24 * 365,
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		Authenticator: m.run,
		PayloadFunc:   m.convertOrgMemberToClaims,
	})
}

type jwtMiddleware struct {
	Db *gorm.DB
}

type jwtLoginVals struct {
	OrganizationID     string `json:"organization_id"`
	Email              string `json:"email"`
	ServiceAccountName string `json:"service_account_name"`
	Password           string `json:"password"`
}

func (m jwtMiddleware) run(ginctx *gin.Context) (interface{}, error) {
	var loginVals jwtLoginVals
	var orgMember dbmodels.IOrganizationMember
	var err error

	if err = ginctx.ShouldBind(&loginVals); err != nil {
		return nil, jwt.ErrMissingLoginValues
	}
	if err = m.validateLoginVals(loginVals); err != nil {
		return nil, err
	}

	if orgMember, err = m.lookupOrgMemberWithLoginVals(loginVals); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("incorrect organization ID or %s", orgMember.IDTypeDisplayName())
		}
		return nil, errors.New("internal database error")
	}

	ok, err := orgMember.Authenticate(loginVals.Password)
	if err != nil {
		return nil, fmt.Errorf("error authenticating organization member: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("incorrect password")
	}

	return orgMember, nil
}

func (m jwtMiddleware) lookupOrgMemberWithLoginVals(loginVals jwtLoginVals) (dbmodels.IOrganizationMember, error) {
	if len(loginVals.Email) > 0 {
		return dbmodels.FindUserByEmail(m.Db, loginVals.OrganizationID,
			loginVals.Email)
	} else if len(loginVals.ServiceAccountName) > 0 {
		return dbmodels.FindServiceAccountByName(m.Db, loginVals.OrganizationID,
			loginVals.ServiceAccountName)
	} else {
		panic("Bug")
	}
}

func (m jwtMiddleware) validateLoginVals(loginVals jwtLoginVals) error {
	if len(loginVals.OrganizationID) == 0 {
		return errors.New("missing organization ID")
	}
	if len(loginVals.Email) == 0 && len(loginVals.ServiceAccountName) == 0 {
		return errors.New("missing login ID")
	}
	if len(loginVals.Email) > 0 && len(loginVals.ServiceAccountName) > 0 {
		return errors.New("only email or service_account_name may be specified, not both")
	}
	if len(loginVals.Password) == 0 {
		return errors.New("missing password")
	}
	return nil
}

func (m jwtMiddleware) convertOrgMemberToClaims(data interface{}) jwt.MapClaims {
	orgMember := data.(dbmodels.IOrganizationMember)
	return jwt.MapClaims{
		jwtOrgIDClaim:         orgMember.GetOrganizationID(),
		jwtOrgMemberTypeClaim: orgMember.Type(),
		jwtOrgMemberIDClaim:   orgMember.ID(),
	}
}

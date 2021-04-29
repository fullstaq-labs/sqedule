package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/fullstaq-labs/sqedule/server/httpapi/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HTTPTestContext ...
type HTTPTestContext struct {
	Db           *gorm.DB
	Engine       *gin.Engine
	HttpCtx      Context
	HttpRecorder *httptest.ResponseRecorder

	Org            dbmodels.Organization
	ServiceAccount dbmodels.ServiceAccount
}

func (ctx HTTPTestContext) ServeHTTP(req *http.Request) {
	ctx.Engine.ServeHTTP(ctx.HttpRecorder, req)
}

// NewRequestWithAuth creates a new Request which is already authenticated
// as `ctx.serviceAccount`.
//
// `body` is either nil, or an object which will be marshalled into JSON
// as the request body.
func (ctx HTTPTestContext) NewRequestWithAuth(method string, url string, body interface{}) (*http.Request, error) {
	var bodyIO io.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		bodyIO = bytes.NewReader(bodyJSON)
	}

	req, err := http.NewRequest(method, url, bodyIO)
	if err == nil {
		SetupHTTPTestAuthentication(req, ctx.Org, ctx.ServiceAccount)
	}
	return req, err
}

// BodyJSON returns the response body as a JSON object.
func (ctx HTTPTestContext) BodyJSON() (gin.H, error) {
	var body gin.H
	err := json.Unmarshal([]byte(ctx.HttpRecorder.Body.String()), &body)
	return body, err
}

// SetupHTTPTestContext ...
func SetupHTTPTestContext() (HTTPTestContext, error) {
	var ctx HTTPTestContext
	var err error

	ctx.Db, err = dbutils.SetupTestDatabase()
	if err != nil {
		return HTTPTestContext{}, err
	}

	ctx.Engine = gin.Default()
	gin.SetMode(gin.DebugMode)

	orgMemberLookupMiddleware := auth.NewOrgMemberLookupMiddleware(ctx.Db, true)
	routingGroup := ctx.Engine.Group("/v1")
	routingGroup.Use(orgMemberLookupMiddleware)

	ctx.HttpCtx = Context{Db: ctx.Db}
	ctx.HttpCtx.InstallRoutes(routingGroup)

	ctx.HttpRecorder = httptest.NewRecorder()

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		ctx.Org, err = dbmodels.CreateMockOrganization(ctx.Db)
		if err != nil {
			return err
		}

		ctx.ServiceAccount, err = dbmodels.CreateMockServiceAccountWithAdminRole(ctx.Db, ctx.Org, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return HTTPTestContext{}, err
	}

	return ctx, nil
}

// SetupHTTPTestAuthentication sets up headers in the given request,
// so that the request is authenticated as the given organization member.
func SetupHTTPTestAuthentication(req *http.Request, organization dbmodels.Organization, orgMember dbmodels.IOrganizationMember) {
	req.Header.Set("TestOrgID", organization.ID)
	req.Header.Set("TestOrgMemberType", string(orgMember.Type()))
	req.Header.Set("TestOrgMemberID", orgMember.ID())
}

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

type HTTPTestContext struct {
	Db *gorm.DB

	Engine        *gin.Engine
	ControllerCtx Context
	Recorder      *httptest.ResponseRecorder

	Org            dbmodels.Organization
	ServiceAccount dbmodels.ServiceAccount
}

func (ctx HTTPTestContext) ServeHTTP(req *http.Request) {
	ctx.Engine.ServeHTTP(ctx.Recorder, req)
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
	err := json.Unmarshal([]byte(ctx.Recorder.Body.String()), &body)
	return body, err
}

func SetupHTTPTestContext(initializer func(ctx *HTTPTestContext, tx *gorm.DB) error) (HTTPTestContext, error) {
	var hctx HTTPTestContext
	var err error

	hctx.Db, err = dbutils.SetupTestDatabase()
	if err != nil {
		return HTTPTestContext{}, err
	}

	gin.SetMode(gin.TestMode)
	hctx.Engine = gin.Default()

	orgMemberLookupMiddleware := auth.NewOrgMemberLookupMiddleware(hctx.Db, true)
	routingGroup := hctx.Engine.Group("/v1")
	routingGroup.Use(orgMemberLookupMiddleware)

	hctx.ControllerCtx = NewContext(hctx.Db)
	hctx.ControllerCtx.InstallRoutes(routingGroup)

	hctx.Recorder = httptest.NewRecorder()

	err = hctx.Db.Transaction(func(tx *gorm.DB) error {
		hctx.Org, err = dbmodels.CreateMockOrganization(tx, nil)
		if err != nil {
			return err
		}

		hctx.ServiceAccount, err = dbmodels.CreateMockServiceAccountWithAdminRole(tx, hctx.Org, nil)
		if err != nil {
			return err
		}

		if initializer != nil {
			err = initializer(&hctx, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return HTTPTestContext{}, err
	}

	return hctx, nil
}

// SetupHTTPTestAuthentication sets up headers in the given request,
// so that the request is authenticated as the given organization member.
func SetupHTTPTestAuthentication(req *http.Request, organization dbmodels.Organization, orgMember dbmodels.IOrganizationMember) {
	req.Header.Set("TestOrgID", organization.ID)
	req.Header.Set("TestOrgMemberType", string(orgMember.Type()))
	req.Header.Set("TestOrgMemberID", orgMember.ID())
}

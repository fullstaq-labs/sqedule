package httpapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HTTPTestContext ...
type HTTPTestContext struct {
	db           *gorm.DB
	engine       *gin.Engine
	httpCtx      Context
	httpRecorder *httptest.ResponseRecorder

	org            dbmodels.Organization
	serviceAccount dbmodels.ServiceAccount
}

func (ctx HTTPTestContext) ServeHTTP(req *http.Request) {
	ctx.engine.ServeHTTP(ctx.httpRecorder, req)
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
		SetupHTTPTestAuthentication(req, ctx.org, ctx.serviceAccount)
	}
	return req, err
}

// BodyJSON returns the response body as a JSON object.
func (ctx HTTPTestContext) BodyJSON() (gin.H, error) {
	var body gin.H
	err := json.Unmarshal([]byte(ctx.httpRecorder.Body.String()), &body)
	return body, err
}

// SetupHTTPTestContext ...
func SetupHTTPTestContext() (HTTPTestContext, error) {
	var ctx HTTPTestContext
	var err error

	ctx.db, err = dbutils.SetupTestDatabase()
	if err != nil {
		return HTTPTestContext{}, err
	}

	ctx.engine = gin.Default()
	gin.SetMode(gin.DebugMode)

	ctx.httpCtx = Context{Db: ctx.db, UseTestAuthentication: true}
	err = ctx.httpCtx.SetupRouter(ctx.engine)
	if err != nil {
		return HTTPTestContext{}, err
	}

	ctx.httpRecorder = httptest.NewRecorder()

	err = ctx.db.Transaction(func(tx *gorm.DB) error {
		ctx.org, err = dbmodels.CreateMockOrganization(ctx.db)
		if err != nil {
			return err
		}

		ctx.serviceAccount, err = dbmodels.CreateMockServiceAccountWithAdminRole(ctx.db, ctx.org, nil)
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

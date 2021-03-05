package httpapi

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateRelease(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}
	app, err := dbmodels.CreateMockApplicationWith1Version(ctx.db, ctx.org)
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("POST", "/v1/applications/"+app.ID+"/releases", gin.H{})
	if !assert.NoError(t, err) {
		return
	}
	ctx.ServeHTTP(req)

	if !assert.Equal(t, 200, ctx.httpRecorder.Code) {
		return
	}
	body, err := ctx.BodyJSON()
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "in_progress", body["state"])
	assert.Nil(t, body["finalized_at"])

	releases, err := dbmodels.FindAllReleases(ctx.db, ctx.org.ID, app.ID)
	if !assert.Equal(t, 1, len(releases)) {
		return
	}
	assert.Equal(t, releasestate.InProgress, releases[0].State)

	bindings, err := dbmodels.FindAllReleaseApprovalRulesetBindings(ctx.db, ctx.org.ID, app.ID, releases[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(bindings))

	var creationEvent dbmodels.ReleaseCreatedEvent
	err = dbutils.CreateFindOperationError(ctx.db.Take(&creationEvent))
	assert.NoError(t, err)

	var creationRecord dbmodels.CreationAuditRecord
	err = dbutils.CreateFindOperationError(ctx.db.Take(&creationRecord))
	if assert.NoError(t, err) {
		assert.False(t, creationRecord.OrganizationMemberIP.Valid)
		assert.Equal(t, ctx.serviceAccount.Name, creationRecord.ServiceAccountName.String)
		if assert.NotNil(t, creationRecord.ReleaseCreatedEventID) {
			assert.Equal(t, creationEvent.ID, *creationRecord.ReleaseCreatedEventID)
		}
	}
}

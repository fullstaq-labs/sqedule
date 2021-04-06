package httpapi

import (
	"fmt"
	"testing"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetRelease(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}

	var app dbmodels.Application
	var release dbmodels.Release
	err = ctx.db.Transaction(func(tx *gorm.DB) error {
		app, err = dbmodels.CreateMockApplicationWith1Version(ctx.db, ctx.org, nil, nil)
		if err != nil {
			return err
		}

		release, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.db, ctx.org, app, nil)
		if err != nil {
			return err
		}

		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.db, ctx.org, "ruleset1", nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.db, ctx.org, release,
			ruleset, *ruleset.LatestMajorVersion, *ruleset.LatestMinorVersion, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s/releases/%d", app.ID, release.ID), nil)
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

	assert.Nil(t, body["application"])
	assert.Equal(t, "in_progress", body["state"])
	assert.Nil(t, body["finalized_at"])
	if !assert.NotEmpty(t, body["approval_ruleset_bindings"]) {
		return
	}

	bindings := body["approval_ruleset_bindings"].([]interface{})
	if !assert.Equal(t, 1, len(bindings)) {
		return
	}

	binding := bindings[0].(map[string]interface{})
	assert.Equal(t, "enforcing", binding["mode"])
	if !assert.NotNil(t, binding["approval_ruleset"]) {
		return
	}

	ruleset := binding["approval_ruleset"].(map[string]interface{})
	assert.Equal(t, "ruleset1", ruleset["id"])
	assert.Equal(t, float64(1), ruleset["major_version_number"])
	assert.Equal(t, float64(1), ruleset["minor_version_number"])
}

func TestCreateRelease(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}
	app, err := dbmodels.CreateMockApplicationWith1Version(ctx.db, ctx.org, nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	_, _, err = dbmodels.CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(ctx.db, ctx.org, app)
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
	fmt.Printf("%#v\n", body)
	assert.Nil(t, body["application"])
	assert.Equal(t, "in_progress", body["state"])
	assert.Nil(t, body["finalized_at"])

	bindingsJSON := body["approval_ruleset_bindings"].([]interface{})
	assert.Equal(t, 2, len(bindingsJSON))

	releases, err := dbmodels.FindAllReleases(ctx.db, ctx.org.ID, app.ID)
	if !assert.Equal(t, 1, len(releases)) {
		return
	}
	assert.Equal(t, releasestate.InProgress, releases[0].State)

	bindings, err := dbmodels.FindAllReleaseApprovalRulesetBindings(ctx.db, ctx.org.ID, app.ID, releases[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bindings))

	var creationEvent dbmodels.ReleaseCreatedEvent
	tx := ctx.db.Take(&creationEvent)
	err = dbutils.CreateFindOperationError(tx)
	assert.NoError(t, err)

	var creationRecord dbmodels.CreationAuditRecord
	tx = ctx.db.Take(&creationRecord)
	err = dbutils.CreateFindOperationError(tx)
	if assert.NoError(t, err) {
		assert.False(t, creationRecord.OrganizationMemberIP.Valid)
		assert.Equal(t, ctx.serviceAccount.Name, creationRecord.ServiceAccountName.String)
		if assert.NotNil(t, creationRecord.ReleaseCreatedEventID) {
			assert.Equal(t, creationEvent.ID, *creationRecord.ReleaseCreatedEventID)
		}
	}

	_, err = dbmodels.FindReleaseBackgroundJob(ctx.db, ctx.org.ID, app.ID, releases[0].ID)
	assert.NoError(t, err)
}

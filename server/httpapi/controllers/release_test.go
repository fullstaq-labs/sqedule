package controllers

import (
	"fmt"
	"testing"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
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
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		app, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
		if err != nil {
			return err
		}

		release, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
		if err != nil {
			return err
		}

		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release,
			ruleset, *ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
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

	if !assert.Equal(t, 200, ctx.HttpRecorder.Code) {
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
	assert.Equal(t, float64(1), ruleset["version_number"])
	assert.Equal(t, float64(1), ruleset["adjustment_number"])
}

func TestCreateRelease(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}
	app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
	if !assert.NoError(t, err) {
		return
	}
	_, _, err = dbmodels.CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(ctx.Db, ctx.Org, app)
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("POST", "/v1/applications/"+app.ID+"/releases", gin.H{})
	if !assert.NoError(t, err) {
		return
	}
	ctx.ServeHTTP(req)

	if !assert.Equal(t, 200, ctx.HttpRecorder.Code) {
		return
	}
	body, err := ctx.BodyJSON()
	if !assert.NoError(t, err) {
		return
	}
	assert.Nil(t, body["application"])
	assert.Equal(t, "in_progress", body["state"])
	assert.Nil(t, body["finalized_at"])

	bindingsJSON := body["approval_ruleset_bindings"].([]interface{})
	assert.Equal(t, 2, len(bindingsJSON))

	releases, err := dbmodels.FindAllReleases(ctx.Db, ctx.Org.ID, app.ID)
	if !assert.Equal(t, 1, len(releases)) {
		return
	}
	assert.Equal(t, releasestate.InProgress, releases[0].State)

	bindings, err := dbmodels.FindAllReleaseApprovalRulesetBindings(ctx.Db, ctx.Org.ID, app.ID, releases[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(bindings))

	var creationEvent dbmodels.ReleaseCreatedEvent
	tx := ctx.Db.Take(&creationEvent)
	err = dbutils.CreateFindOperationError(tx)
	assert.NoError(t, err)

	var creationRecord dbmodels.CreationAuditRecord
	tx = ctx.Db.Take(&creationRecord)
	err = dbutils.CreateFindOperationError(tx)
	if assert.NoError(t, err) {
		assert.False(t, creationRecord.OrganizationMemberIP.Valid)
		assert.Equal(t, ctx.ServiceAccount.Name, creationRecord.ServiceAccountName.String)
		if assert.NotNil(t, creationRecord.ReleaseCreatedEventID) {
			assert.Equal(t, creationEvent.ID, *creationRecord.ReleaseCreatedEventID)
		}
	}

	_, err = dbmodels.FindReleaseBackgroundJob(ctx.Db, ctx.Org.ID, app.ID, releases[0].ID)
	assert.NoError(t, err)
}

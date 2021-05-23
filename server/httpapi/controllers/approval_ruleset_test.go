package controllers

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetAllApprovalRulesets(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}

	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
		if err != nil {
			return err
		}

		app1, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
			func(app *dbmodels.Application) {
				app.ID = "app1"
			},
			func(adjustment *dbmodels.ApplicationAdjustment) {
				adjustment.DisplayName = "App 1"
			})
		if err != nil {
			return err
		}
		app2, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
			func(app *dbmodels.Application) {
				app.ID = "app2"
			},
			func(adjustment *dbmodels.ApplicationAdjustment) {
				adjustment.DisplayName = "App 2"
			})
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app1, ruleset, nil)
		if err != nil {
			return err
		}
		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app2, ruleset, nil)
		if err != nil {
			return err
		}

		release1, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app1, nil)
		if err != nil {
			return err
		}
		release2, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app2, nil)
		if err != nil {
			return err
		}
		release3, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app2, nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release1, ruleset, *ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
		if err != nil {
			return err
		}
		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release2, ruleset, *ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
		if err != nil {
			return err
		}
		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release3, ruleset, *ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("GET", "/v1/approval-rulesets", nil)
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

	if !assert.NotNil(t, body["items"]) {
		return
	}
	items := body["items"].([]interface{})
	if !assert.Equal(t, 1, len(items)) {
		return
	}

	ruleset := items[0].(map[string]interface{})
	assert.Equal(t, "ruleset1", ruleset["id"])
	assert.Equal(t, float64(1), ruleset["version_number"])
	assert.Equal(t, float64(1), ruleset["adjustment_number"])
	assert.Equal(t, float64(2), ruleset["num_bound_applications"])
	assert.Equal(t, float64(3), ruleset["num_bound_releases"])
}

func TestGetApprovalRuleset(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}

	var mockRelease dbmodels.Release
	var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
		if err != nil {
			return err
		}

		app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, ruleset, nil)
		if err != nil {
			return err
		}

		mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, mockRelease, ruleset,
			*ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
		if err != nil {
			return err
		}

		mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
			ruleset.LatestVersion.ID, *ruleset.LatestAdjustment, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("GET", "/v1/approval-rulesets/ruleset1", nil)
	if !assert.NoError(t, err) {
		return
	}
	ctx.ServeHTTP(req)

	if !assert.Equal(t, 200, ctx.HttpRecorder.Code) {
		return
	}
	ruleset, err := ctx.BodyJSON()
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "ruleset1", ruleset["id"])
	assert.Equal(t, float64(1), ruleset["version_number"])
	assert.Equal(t, float64(1), ruleset["adjustment_number"])

	if !assert.NotEmpty(t, ruleset["application_approval_ruleset_bindings"]) {
		return
	}
	appBindings := ruleset["application_approval_ruleset_bindings"].([]interface{})
	if !assert.Equal(t, 1, len(appBindings)) {
		return
	}
	appBinding := appBindings[0].(map[string]interface{})
	assert.Equal(t, "enforcing", appBinding["mode"])
	if !assert.NotNil(t, appBinding["application"]) {
		return
	}
	app := appBinding["application"].(map[string]interface{})
	assert.Equal(t, "app1", app["id"])

	if !assert.NotEmpty(t, ruleset["release_approval_ruleset_bindings"]) {
		return
	}
	releaseBindings := ruleset["release_approval_ruleset_bindings"].([]interface{})
	if !assert.Equal(t, 1, len(releaseBindings)) {
		return
	}
	releaseBinding := releaseBindings[0].(map[string]interface{})
	assert.Equal(t, "enforcing", releaseBinding["mode"])
	if !assert.NotNil(t, releaseBinding["release"]) {
		return
	}
	release := releaseBinding["release"].(map[string]interface{})
	assert.Equal(t, float64(mockRelease.ID), release["id"])
	if !assert.NotNil(t, release["application"]) {
		return
	}
	app = release["application"].(map[string]interface{})
	assert.Equal(t, "app1", app["id"])

	if !assert.NotEmpty(t, ruleset["approval_rules"]) {
		return
	}
	rules := ruleset["approval_rules"].([]interface{})
	if !assert.Equal(t, 1, len(rules)) {
		return
	}
	rule := rules[0].(map[string]interface{})
	assert.Equal(t, float64(mockScheduleApprovalRule.ID), rule["id"])
	assert.Equal(t, mockScheduleApprovalRule.BeginTime.String, rule["begin_time"])
}

package httpapi

import (
	"fmt"
	"testing"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetAllApplications(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}

	var app1, app2 dbmodels.Application
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		app1, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
			func(app *dbmodels.Application) {
				app.CreatedAt = time.Date(2021, 3, 8, 12, 0, 0, 0, time.Local)
			},
			nil)
		if err != nil {
			return err
		}

		app2, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
			func(app *dbmodels.Application) {
				app.ID = "app2"
				app.CreatedAt = time.Date(2021, 3, 8, 11, 0, 0, 0, time.Local)
			},
			func(minorVersion *dbmodels.ApplicationMinorVersion) {
				minorVersion.DisplayName = "App 2"
			})
		if err != nil {
			return err
		}

		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app1,
			ruleset, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications"), nil)
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

	if !assert.NotEmpty(t, body["items"]) {
		return
	}
	items := body["items"].([]interface{})
	if !assert.Equal(t, 2, len(items)) {
		return
	}

	item1 := items[0].(map[string]interface{})
	assert.Equal(t, app1.ID, item1["id"])
	assert.Equal(t, float64(1), item1["major_version_number"])
	assert.Equal(t, float64(1), item1["minor_version_number"])
	assert.Equal(t, app1.LatestMinorVersion.DisplayName, item1["display_name"])
	assert.Equal(t, app1.LatestMinorVersion.Enabled, item1["enabled"])
	assert.NotNil(t, item1["created_at"])
	assert.NotNil(t, item1["updated_at"])
	assert.Empty(t, item1["approval_ruleset_bindings"])

	item2 := items[1].(map[string]interface{})
	assert.Equal(t, app2.ID, item2["id"])
	assert.Equal(t, float64(1), item2["major_version_number"])
	assert.Equal(t, float64(1), item2["minor_version_number"])
	assert.Equal(t, app2.LatestMinorVersion.DisplayName, item2["display_name"])
	assert.Equal(t, app2.LatestMinorVersion.Enabled, item2["enabled"])
	assert.NotNil(t, item2["created_at"])
	assert.NotNil(t, item2["updated_at"])
	assert.Empty(t, item2["approval_ruleset_bindings"])
}

func TestGetApplication(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}

	var app dbmodels.Application
	err = ctx.Db.Transaction(func(tx *gorm.DB) error {
		app, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
		if err != nil {
			return err
		}

		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app,
			ruleset, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if !assert.NoError(t, err) {
		return
	}

	req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s", app.ID), nil)
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

	assert.Equal(t, app.ID, body["id"])
	assert.Equal(t, float64(1), body["major_version_number"])
	assert.Equal(t, float64(1), body["minor_version_number"])
	assert.Equal(t, app.LatestMinorVersion.DisplayName, body["display_name"])
	assert.Equal(t, app.LatestMinorVersion.Enabled, body["enabled"])
	assert.NotNil(t, body["created_at"])
	assert.NotNil(t, body["updated_at"])
	if !assert.NotEmpty(t, body["approval_ruleset_bindings"]) {
		return
	}

	bindings := body["approval_ruleset_bindings"].([]interface{})
	if !assert.Equal(t, 1, len(bindings)) {
		return
	}

	binding := bindings[0].(map[string]interface{})
	assert.Equal(t, "enforcing", binding["mode"])
	assert.Equal(t, float64(1), binding["major_version_number"])
	assert.Equal(t, float64(1), binding["minor_version_number"])
	if !assert.NotNil(t, binding["approval_ruleset"]) {
		return
	}

	ruleset := binding["approval_ruleset"].(map[string]interface{})
	assert.Equal(t, "ruleset1", ruleset["id"])
	assert.Equal(t, float64(1), ruleset["major_version_number"])
	assert.Equal(t, float64(1), ruleset["minor_version_number"])
}

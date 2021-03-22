package httpapi

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestGetAllApprovalRulesets(t *testing.T) {
	ctx, err := SetupHTTPTestContext()
	if !assert.NoError(t, err) {
		return
	}

	err = ctx.db.Transaction(func(tx *gorm.DB) error {
		ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.db, ctx.org, "ruleset1", nil)
		if err != nil {
			return err
		}

		app1, err := dbmodels.CreateMockApplicationWith1Version(ctx.db, ctx.org,
			func(app *dbmodels.Application) {
				app.ID = "app1"
			},
			func(minorVersion *dbmodels.ApplicationMinorVersion) {
				minorVersion.DisplayName = "App 1"
			})
		if err != nil {
			return err
		}
		app2, err := dbmodels.CreateMockApplicationWith1Version(ctx.db, ctx.org,
			func(app *dbmodels.Application) {
				app.ID = "app2"
			},
			func(minorVersion *dbmodels.ApplicationMinorVersion) {
				minorVersion.DisplayName = "App 2"
			})
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.db, ctx.org, app1, ruleset, nil)
		if err != nil {
			return err
		}
		_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.db, ctx.org, app2, ruleset, nil)
		if err != nil {
			return err
		}

		release1, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.db, ctx.org, app1, nil)
		if err != nil {
			return err
		}
		release2, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.db, ctx.org, app2, nil)
		if err != nil {
			return err
		}
		release3, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.db, ctx.org, app2, nil)
		if err != nil {
			return err
		}

		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.db, ctx.org, release1, ruleset, *ruleset.LatestMajorVersion, *ruleset.LatestMinorVersion, nil)
		if err != nil {
			return err
		}
		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.db, ctx.org, release2, ruleset, *ruleset.LatestMajorVersion, *ruleset.LatestMinorVersion, nil)
		if err != nil {
			return err
		}
		_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.db, ctx.org, release3, ruleset, *ruleset.LatestMajorVersion, *ruleset.LatestMinorVersion, nil)
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

	if !assert.Equal(t, 200, ctx.httpRecorder.Code) {
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
	assert.Equal(t, float64(1), ruleset["major_version_number"])
	assert.Equal(t, float64(1), ruleset["minor_version_number"])
	assert.Equal(t, float64(2), ruleset["num_bound_applications"])
	assert.Equal(t, float64(3), ruleset["num_bound_releases"])
}

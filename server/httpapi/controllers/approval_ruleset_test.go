package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"gorm.io/gorm"
)

var _ = Describe("approval-ruleset API", func() {
	var ctx HTTPTestContext
	var err error

	BeforeEach(func() {
		ctx, err = SetupHTTPTestContext()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("GET /approval-rulesets", func() {
		var body gin.H

		BeforeEach(func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app1, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
					func(app *dbmodels.Application) {
						app.ID = "app1"
					},
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.DisplayName = "App 1"
					})
				Expect(err).ToNot(HaveOccurred())

				app2, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
					func(app *dbmodels.Application) {
						app.ID = "app2"
					},
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.DisplayName = "App 2"
					})
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app1, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app2, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				release1, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app1, nil)
				Expect(err).ToNot(HaveOccurred())
				release2, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app2, nil)
				Expect(err).ToNot(HaveOccurred())
				release3, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app2, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release1, ruleset,
					*ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release2, ruleset,
					*ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release3, ruleset,
					*ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", "/v1/approval-rulesets", nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs all approval rulesets", func() {
			Expect(body["items"]).NotTo(BeNil())
			items := body["items"].([]interface{})
			Expect(items).To(HaveLen(1))

			ruleset := items[0].(map[string]interface{})
			Expect(ruleset["id"]).To(Equal("ruleset1"))
			Expect(ruleset["version_number"]).To(Equal(float64(1)))
			Expect(ruleset["adjustment_number"]).To(Equal(float64(1)))
			Expect(ruleset["num_bound_applications"]).To(Equal(float64(2)))
			Expect(ruleset["num_bound_releases"]).To(Equal(float64(3)))
		})
	})

	Describe("GET /approval-rulesets/:id", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule
		var ruleset gin.H

		BeforeEach(func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, mockRelease, ruleset,
					*ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
					ruleset.LatestVersion.ID, *ruleset.LatestAdjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", "/v1/approval-rulesets/ruleset1", nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)
			Expect(ctx.HttpRecorder.Code).To(Equal(200))

			ruleset, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs the latest version", func() {
			Expect(ruleset["id"]).To(Equal("ruleset1"))
			Expect(ruleset["version_number"]).To(Equal(float64(1)))
			Expect(ruleset["adjustment_number"]).To(Equal(float64(1)))
		})

		It("outputs application bindings", func() {
			Expect(ruleset["application_approval_ruleset_bindings"]).ToNot(BeEmpty())
			appBindings := ruleset["application_approval_ruleset_bindings"].([]interface{})
			Expect(appBindings).To(HaveLen(1))
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding["mode"]).To(Equal("enforcing"))
			app := appBinding["application"].(map[string]interface{})
			Expect(app["id"]).To(Equal("app1"))
		})

		It("outputs release bindings", func() {
			Expect(ruleset["release_approval_ruleset_bindings"]).ToNot(BeEmpty())
			releaseBindings := ruleset["release_approval_ruleset_bindings"].([]interface{})
			Expect(releaseBindings).To(HaveLen(1))
			releaseBinding := releaseBindings[0].(map[string]interface{})
			Expect(releaseBinding["mode"]).To(Equal("enforcing"))
			Expect(releaseBinding["release"]).ToNot(BeNil())
			release := releaseBinding["release"].(map[string]interface{})
			Expect(release["id"]).To(Equal(float64(mockRelease.ID)))
			Expect(release["application"]).ToNot(BeNil())
			app := release["application"].(map[string]interface{})
			Expect(app["id"]).To(Equal("app1"))
		})

		It("outputs rules", func() {
			Expect(ruleset["approval_rules"]).ToNot(BeEmpty())
			rules := ruleset["approval_rules"].([]interface{})
			Expect(rules).To(HaveLen(1))
			rule := rules[0].(map[string]interface{})
			Expect(rule["id"]).To(Equal(float64(mockScheduleApprovalRule.ID)))
			Expect(rule["begin_time"]).To(Equal(mockScheduleApprovalRule.BeginTime.String))
		})
	})
})

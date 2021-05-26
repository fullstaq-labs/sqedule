package controllers

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

var _ = Describe("approval-ruleset API", func() {
	var ctx HTTPTestContext
	var err error

	BeforeEach(func() {
		ctx, err = SetupHTTPTestContext()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("POST /approval-rulesets", func() {
		var body gin.H

		Setup := func(proposalState string) {
			input := gin.H{
				"id": "ruleset1",
				"version": gin.H{
					"display_name": "Ruleset 1",
					"approval_rules": []gin.H{
						{
							"type":       "schedule",
							"begin_time": "1:00",
							"end_time":   "2:00",
						},
					},
				},
			}
			if len(proposalState) > 0 {
				input["version"].(gin.H)["proposal_state"] = proposalState
			}

			req, err := ctx.NewRequestWithAuth("POST", "/v1/approval-rulesets", input)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(201))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		}

		It("outputs the created approval ruleset", func() {
			Setup("")

			Expect(body).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(body["version"]).ToNot(BeNil())

			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset 1"))
		})

		It("creates a new approval ruleset", func() {
			Setup("")

			var ruleset dbmodels.ApprovalRuleset
			tx := ctx.Db.Take(&ruleset)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

			Expect(ruleset.ID).To(Equal("ruleset1"))

			var adjustment dbmodels.ApprovalRulesetAdjustment
			tx = ctx.Db.Take(&adjustment)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(adjustment.DisplayName).To(Equal("Ruleset 1"))

			var rule dbmodels.ScheduleApprovalRule
			tx = ctx.Db.Take(&rule)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(rule.BeginTime.String).To(Equal("1:00"))
			Expect(rule.EndTime.String).To(Equal("2:00"))
		})

		It("creates a draft proposal by default", func() {
			Setup("")

			Expect(body["version"]).ToNot(BeNil())
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
			Expect(version["version_number"]).To(BeNil())
			Expect(version["approved_at"]).To(BeNil())

			var adjustment dbmodels.ApprovalRulesetAdjustment
			tx := ctx.Db.Take(&adjustment)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(adjustment.ReviewState).To(Equal(reviewstate.Draft))
		})

		It("creates a draft proposal if proposal_state is draft", func() {
			Setup("draft")

			Expect(body["version"]).ToNot(BeNil())
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
			Expect(version["version_number"]).To(BeNil())
			Expect(version["approved_at"]).To(BeNil())

			var adjustment dbmodels.ApprovalRulesetAdjustment
			tx := ctx.Db.Take(&adjustment)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(adjustment.ReviewState).To(Equal(reviewstate.Draft))
		})

		It("submits the version for approval if proposal_state is final", func() {
			Setup("final")

			Expect(body["version"]).ToNot(BeNil())
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("version_state", "approved"))
			Expect(version["version_number"]).To(BeNumerically("==", 1))
			Expect(version["approved_at"]).ToNot(BeNil())

			var adjustment dbmodels.ApprovalRulesetAdjustment
			tx := ctx.Db.Take(&adjustment)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(adjustment.ReviewState).To(Equal(reviewstate.Approved))
		})
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
			Expect(body["items"]).To(HaveLen(1))

			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})
			Expect(ruleset["id"]).To(Equal("ruleset1"))
			Expect(ruleset["latest_approved_version"]).ToNot(BeNil())
			Expect(ruleset["num_bound_applications"]).To(BeNumerically("==", 2))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version["version_number"]).To(BeNumerically("==", 1))
			Expect(version["num_bound_releases"]).To(BeNumerically("==", 3))
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
			Expect(ruleset["latest_approved_version"]).ToNot(BeNil())
			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version["version_number"]).To(BeNumerically("==", 1))
		})

		It("outputs application bindings", func() {
			Expect(ruleset["application_approval_ruleset_bindings"]).To(HaveLen(1))
			appBindings := ruleset["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding["mode"]).To(Equal("enforcing"))
			app := appBinding["application"].(map[string]interface{})
			Expect(app["id"]).To(Equal("app1"))
		})

		It("outputs release bindings", func() {
			Expect(ruleset["latest_approved_version"]).ToNot(BeNil())
			version := ruleset["latest_approved_version"].(map[string]interface{})

			Expect(version["release_approval_ruleset_bindings"]).To(HaveLen(1))
			releaseBindings := version["release_approval_ruleset_bindings"].([]interface{})
			releaseBinding := releaseBindings[0].(map[string]interface{})
			Expect(releaseBinding["mode"]).To(Equal("enforcing"))
			Expect(releaseBinding["release"]).ToNot(BeNil())

			release := releaseBinding["release"].(map[string]interface{})
			Expect(release["id"]).To(BeNumerically("==", mockRelease.ID))
			Expect(release["application"]).ToNot(BeNil())

			app := release["application"].(map[string]interface{})
			Expect(app["id"]).To(Equal("app1"))
		})

		It("outputs rules", func() {
			Expect(ruleset["latest_approved_version"]).ToNot(BeNil())
			version := ruleset["latest_approved_version"].(map[string]interface{})

			Expect(version["approval_rules"]).To(HaveLen(1))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule["id"]).To(BeNumerically("==", mockScheduleApprovalRule.ID))
			Expect(rule["begin_time"]).To(Equal(mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id", func() {
		It("patches an approval ruleset's unversioned data", func() {
			_, err = dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("PATCH", "/v1/approval-rulesets/ruleset1", gin.H{
				"id": "ruleset2",
			})
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)
			Expect(ctx.HttpRecorder.Code).To(Equal(200))

			body, err := ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
			Expect(body).To(HaveKeyWithValue("id", "ruleset2"))

			var ruleset dbmodels.ApprovalRuleset
			tx := ctx.Db.Take(&ruleset)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(ruleset.ID).To(Equal("ruleset2"))
		})

		Describe("upon patching an approval ruleset's versioned data", func() {
			var body gin.H
			var mockRuleset dbmodels.ApprovalRuleset

			Setup := func(proposalState string) {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					mockRuleset, err = dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, mockRuleset, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
						mockRuleset.LatestVersion.ID, *mockRuleset.LatestAdjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
						mockRuleset.LatestVersion.ID, *mockRuleset.LatestAdjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					return nil
				})
				Expect(err).ToNot(HaveOccurred())

				input := gin.H{
					"version": gin.H{
						"display_name": "Ruleset 2",
						"approval_rules": []gin.H{
							{
								"type":       "schedule",
								"begin_time": "1:00",
								"end_time":   "2:00",
							},
						},
					},
				}
				if len(proposalState) > 0 {
					input["version"].(gin.H)["proposal_state"] = proposalState
				}

				req, err := ctx.NewRequestWithAuth("PATCH", "/v1/approval-rulesets/ruleset1", input)
				Expect(err).ToNot(HaveOccurred())
				ctx.ServeHTTP(req)

				Expect(ctx.HttpRecorder.Code).To(Equal(200))
				body, err = ctx.BodyJSON()
				Expect(err).ToNot(HaveOccurred())
			}

			It("outputs the created approval ruleset version", func() {
				Setup("")

				Expect(body["version"]).ToNot(BeNil())
				version := body["version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("display_name", "Ruleset 2"))

				rules := version["approval_rules"].([]interface{})
				Expect(rules).To(HaveLen(1))
				Expect(rules[0]).To(HaveKeyWithValue("type", "schedule"))
				Expect(rules[0]).To(HaveKeyWithValue("begin_time", "1:00"))
				Expect(rules[0]).To(HaveKeyWithValue("end_time", "2:00"))
			})

			It("creates a draft proposal by default", func() {
				Setup("")

				Expect(body["version"]).ToNot(BeNil())
				version := body["version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
				Expect(version["version_number"]).To(BeNil())
				Expect(version["approved_at"]).To(BeNil())

				var adjustment dbmodels.ApprovalRulesetAdjustment
				tx := ctx.Db.Where("review_state = 'draft'").Take(&adjustment)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			})

			It("creates a draft proposal if proposal_state is draft", func() {
				Setup("draft")

				Expect(body["version"]).ToNot(BeNil())
				version := body["version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("version_state", "proposal"))
				Expect(version["version_number"]).To(BeNil())
				Expect(version["approved_at"]).To(BeNil())

				var adjustment dbmodels.ApprovalRulesetAdjustment
				tx := ctx.Db.Where("review_state = 'draft'").Take(&adjustment)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			})

			It("submits the version for approval if proposal_state is final", func() {
				Setup("final")

				Expect(body["version"]).ToNot(BeNil())
				version := body["version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("version_state", "approved"))
				Expect(version["version_number"]).To(BeNumerically("==", 2))
				Expect(version["approved_at"]).ToNot(BeNil())

				err = dbmodels.LoadApprovalRulesetsLatestVersions(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRuleset{&mockRuleset})
				Expect(err).ToNot(HaveOccurred())
				Expect(mockRuleset.LatestVersion).ToNot(BeNil())
				Expect(*mockRuleset.LatestVersion.VersionNumber).To(BeNumerically("==", 2))
				Expect(mockRuleset.LatestAdjustment.ReviewState).To(Equal(reviewstate.Approved))
			})

			It("creates new ApprovalRule objects rather than modifying existing ones in place", func() {
				Setup("")

				var count int64
				err = ctx.Db.Model(dbmodels.ScheduleApprovalRule{}).Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(BeNumerically("==", 3))
			})
		})
	})
})

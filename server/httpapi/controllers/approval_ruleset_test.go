package controllers

import (
	"reflect"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
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
		includedTestCtx := IncludeReviewableCreateTest(ReviewableCreateTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/approval-rulesets",
			UnversionedInput: gin.H{
				"id": "ruleset1",
			},
			VersionedInput: gin.H{
				"display_name": "Ruleset 1",
				"approval_rules": []gin.H{
					{
						"type":       "schedule",
						"begin_time": "1:00",
						"end_time":   "2:00",
					},
				},
			},
			ResourceType:   reflect.TypeOf(dbmodels.ApprovalRuleset{}),
			AdjustmentType: reflect.TypeOf(dbmodels.ApprovalRulesetAdjustment{}),
			GetPrimaryKey: func(resource interface{}) interface{} {
				return resource.(*dbmodels.ApprovalRuleset).ID
			},
			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
			GetVersionedField: func(adjustment interface{}) interface{} {
				return adjustment.(*dbmodels.ApprovalRulesetAdjustment).DisplayName
			},
			VersionedFieldJSONFieldName: "display_name",
			VersionedFieldInitialValue:  "Ruleset 1",
		})

		It("creates rule objects", func() {
			includedTestCtx.MakeRequest("", 201)

			var rule dbmodels.ScheduleApprovalRule
			tx := ctx.Db.Take(&rule)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(rule.BeginTime.String).To(Equal("1:00"))
			Expect(rule.EndTime.String).To(Equal("2:00"))
		})
	})

	Describe("GET /approval-rulesets", func() {
		Setup := func() (ruleset dbmodels.ApprovalRuleset) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err = dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())
				return nil
			})
			Expect(err).ToNot(HaveOccurred())
			return ruleset
		}

		SetupAssociations := func(ruleset dbmodels.ApprovalRuleset) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
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
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release2, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release3, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadAllTest(ReviewableReadAllTestOptions{
			HTTPTestCtx:             &ctx,
			Path:                    "/v1/approval-rulesets",
			Setup:                   func() { Setup() },
			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
		})

		It("outputs the number of bound applications and releases", func() {
			ruleset := Setup()
			SetupAssociations(ruleset)
			body := includedTestCtx.MakeRequest()

			Expect(body["items"]).To(HaveLen(1))

			items := body["items"].([]interface{})
			rulesetJSON := items[0].(map[string]interface{})
			Expect(rulesetJSON["num_bound_applications"]).To(BeNumerically("==", 2))

			version := rulesetJSON["latest_approved_version"].(map[string]interface{})
			Expect(version["num_bound_releases"]).To(BeNumerically("==", 3))
		})
	})

	Describe("GET /approval-rulesets/:id", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func() {
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
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
					ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadTest(ReviewableReadTestOptions{
			HTTPTestCtx:             &ctx,
			Path:                    "/v1/approval-rulesets/ruleset1",
			Setup:                   Setup,
			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
		})

		It("outputs application bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body["application_approval_ruleset_bindings"]).To(HaveLen(1))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding["mode"]).To(Equal("enforcing"))
			app := appBinding["application"].(map[string]interface{})
			Expect(app["id"]).To(Equal("app1"))
		})

		It("outputs release bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body["latest_approved_version"]).ToNot(BeNil())
			version := body["latest_approved_version"].(map[string]interface{})

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

		It("outputs approval rules", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body["latest_approved_version"]).ToNot(BeNil())
			version := body["latest_approved_version"].(map[string]interface{})

			Expect(version["approval_rules"]).To(HaveLen(1))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule["id"]).To(BeNumerically("==", mockScheduleApprovalRule.ID))
			Expect(rule["begin_time"]).To(Equal(mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule
		var body gin.H

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
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
					ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("PATCH", "/v1/approval-rulesets/ruleset1", gin.H{"id": "ruleset2"})
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs application bindings", func() {
			Expect(body["application_approval_ruleset_bindings"]).To(HaveLen(1))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding["mode"]).To(Equal("enforcing"))
			app := appBinding["application"].(map[string]interface{})
			Expect(app["id"]).To(Equal("app1"))
		})

		It("outputs release bindings", func() {
			Expect(body["version"]).ToNot(BeNil())
			version := body["version"].(map[string]interface{})

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

		It("outputs approval rules", func() {
			Expect(body["version"]).ToNot(BeNil())
			version := body["version"].(map[string]interface{})

			Expect(version["approval_rules"]).To(HaveLen(1))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule["id"]).To(BeNumerically("==", mockScheduleApprovalRule.ID))
			Expect(rule["begin_time"]).To(Equal(mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id", func() {
		Describe("upon patching an approval ruleset's unversioned data", func() {
			Setup := func() {
				_, err = dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())
			}

			IncludeReviewableUpdateUnversionedDataTest(ReviewableUpdateUnversionedDataTestOptions{
				HTTPTestCtx:      &ctx,
				Path:             "/v1/approval-rulesets/ruleset1",
				Setup:            Setup,
				UnversionedInput: gin.H{"id": "ruleset2"},
				ResourceType:     reflect.TypeOf(dbmodels.ApprovalRuleset{}),
				GetPrimaryKey: func(resource interface{}) interface{} {
					return resource.(*dbmodels.ApprovalRuleset).ID
				},
				PrimaryKeyJSONFieldName: "id",
				PrimaryKeyUpdatedValue:  "ruleset2",
			})
		})

		Describe("upon patching an approval ruleset's versioned data", func() {
			var mockRuleset dbmodels.ApprovalRuleset

			Setup := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					mockRuleset, err = dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, mockRuleset, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
						mockRuleset.Version.ID, *mockRuleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
						mockRuleset.Version.ID, *mockRuleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					return nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			includedTestCtx := IncludeReviewableUpdateVersionedDataTest(ReviewableUpdateVersionedDataTestOptions{
				HTTPTestCtx: &ctx,
				Path:        "/v1/approval-rulesets/ruleset1",
				Setup:       Setup,
				VersionedInput: gin.H{
					"display_name": "Ruleset 2",
					"approval_rules": []gin.H{
						{
							"type":       "schedule",
							"begin_time": "1:00",
							"end_time":   "2:00",
						},
					},
				},
				AdjustmentType: reflect.TypeOf(dbmodels.ApprovalRulesetAdjustment{}),
				GetLatestResourceVersionAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
					err := dbmodels.LoadApprovalRulesetsLatestVersionsAndAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRuleset{&mockRuleset})
					Expect(err).ToNot(HaveOccurred())
					return mockRuleset.Version, mockRuleset.Version.Adjustment
				},
				VersionedFieldJSONFieldName: "display_name",
				VersionedFieldUpdatedValue:  "Ruleset 2",
			})

			It("outputs the new version's approval rules", func() {
				Setup()
				body := includedTestCtx.MakeRequest("", 200)

				version := body["version"].(map[string]interface{})
				rules := version["approval_rules"].([]interface{})
				Expect(rules).To(HaveLen(1))
				Expect(rules[0]).To(HaveKeyWithValue("type", "schedule"))
				Expect(rules[0]).To(HaveKeyWithValue("begin_time", "1:00"))
				Expect(rules[0]).To(HaveKeyWithValue("end_time", "2:00"))
			})

			It("creates new ApprovalRule objects rather than modifying existing ones in-place", func() {
				Setup()
				includedTestCtx.MakeRequest("", 200)

				var count int64
				err = ctx.Db.Model(dbmodels.ScheduleApprovalRule{}).Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				Expect(count).To(BeNumerically("==", 3))
			})
		})
	})
})

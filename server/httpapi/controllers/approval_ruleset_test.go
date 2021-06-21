package controllers

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"

	"github.com/fullstaq-labs/sqedule/lib"
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
				ruleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
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

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, release1, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, release2, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, release3, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadResourcesTest(ReviewableReadResourcesTestOptions{
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

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))

			items := body["items"].([]interface{})
			rulesetJSON := items[0].(map[string]interface{})
			Expect(rulesetJSON).To(HaveKeyWithValue("num_bound_applications", BeNumerically("==", 2)))

			version := rulesetJSON["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("num_bound_releases", BeNumerically("==", 3)))
		})
	})

	Describe("GET /approval-rulesets/:id", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, mockRelease, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
					ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadResourceTest(ReviewableReadResourceTestOptions{
			HTTPTestCtx:             &ctx,
			Path:                    "/v1/approval-rulesets/ruleset1",
			Setup:                   Setup,
			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
		})

		It("outputs application bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeEmpty())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs release bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))
			version := body["latest_approved_version"].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("release_approval_ruleset_bindings", HaveLen(1)))
			releaseBindings := version["release_approval_ruleset_bindings"].([]interface{})
			releaseBinding := releaseBindings[0].(map[string]interface{})
			Expect(releaseBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(releaseBinding).To(HaveKeyWithValue("release", Not(BeEmpty())))

			release := releaseBinding["release"].(map[string]interface{})
			Expect(release).To(HaveKeyWithValue("id", BeNumerically("==", mockRelease.ID)))
			Expect(release).To(HaveKeyWithValue("application", Not(BeEmpty())))

			app := release["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("latest_approved_version", Not(BeEmpty())))
			version := body["latest_approved_version"].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule
		var body gin.H

		BeforeEach(func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, mockRelease, ruleset,
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
			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeEmpty())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs release bindings", func() {
			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})

			Expect(version["release_approval_ruleset_bindings"]).To(HaveLen(1))
			releaseBindings := version["release_approval_ruleset_bindings"].([]interface{})
			releaseBinding := releaseBindings[0].(map[string]interface{})
			Expect(releaseBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(releaseBinding).To(HaveKeyWithValue("release", Not(BeEmpty())))

			release := releaseBinding["release"].(map[string]interface{})
			Expect(release).To(HaveKeyWithValue("id", BeNumerically("==", mockRelease.ID)))
			Expect(release).To(HaveKeyWithValue("application", Not(BeEmpty())))

			app := release["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id", func() {
		Describe("upon patching unversioned data", func() {
			Setup := func() {
				_, err = dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
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

		Describe("upon patching versioned data", func() {
			var mockRuleset dbmodels.ApprovalRuleset

			Setup := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					mockRuleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
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

				Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
				version := body["version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
				rules := version["approval_rules"].([]interface{})
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

	Describe("GET /approval-rulesets/:id/versions", func() {
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func(approved bool) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRuleset(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				if approved {
					// Create a ruleset with 3 versions
					rulesetVersion1, err := dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, rulesetVersion1, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					// We deliberately create version 3 out of order so that we test
					// whether the versions are outputted in order.

					rulesetVersion3, err := dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, lib.NewUint32Ptr(3), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, rulesetVersion3, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					rulesetVersion2, err := dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, lib.NewUint32Ptr(2), nil)
					Expect(err).ToNot(HaveOccurred())
					rulesetVersion2Adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, rulesetVersion2, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					ruleset.Version = &rulesetVersion3
					ruleset.Version.Adjustment = &rulesetVersion2Adjustment

					app, err := dbmodels.CreateMockApplication(ctx.Db, ctx.Org, nil)
					Expect(err).ToNot(HaveOccurred())

					release1, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
					Expect(err).ToNot(HaveOccurred())
					release2, err := dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, release1,
						ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, release2,
						ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
						ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					proposal, err := dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, proposal, 1,
						func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
							adjustment.ReviewState = reviewstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadVersionsTest(ReviewableReadVersionsTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/approval-rulesets/ruleset1/versions",
			Setup:       Setup,
		})

		It("outputs the number of bound releases", func() {
			Setup(true)
			body := includedTestCtx.MakeRequest()

			Expect(body["items"]).To(HaveLen(3))

			items := body["items"].([]interface{})
			version := items[0].(map[string]interface{})
			Expect(version["num_bound_releases"]).To(BeNumerically("==", 2))
		})

		It("outputs approval rules", func() {
			Setup(true)
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", Not(BeEmpty())))
			items := body["items"].([]interface{})
			version := items[0].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("GET /approval-rulesets/:id/versions/:version_number", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(ctx.Db, ctx.Org, mockRelease,
					ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
					ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadVersionTest(ReviewableReadVersionTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/approval-rulesets/ruleset1/versions/1",
			Setup:       Setup,

			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
		})

		It("outputs application bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeEmpty())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs release bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
			version := body["version"].(map[string]interface{})

			Expect(version["release_approval_ruleset_bindings"]).To(HaveLen(1))
			releaseBindings := version["release_approval_ruleset_bindings"].([]interface{})
			releaseBinding := releaseBindings[0].(map[string]interface{})
			Expect(releaseBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(releaseBinding).To(HaveKeyWithValue("release", Not(BeEmpty())))

			release := releaseBinding["release"].(map[string]interface{})
			Expect(release).To(HaveKeyWithValue("id", BeNumerically("==", mockRelease.ID)))
			Expect(release).To(HaveKeyWithValue("application", Not(BeEmpty())))

			app := release["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
			version := body["version"].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("GET /approval-rulesets/:id/proposals", func() {
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func(approved bool) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRuleset(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				if approved {
					version, err := dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, version, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					proposal, err := dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, proposal, 1,
						func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
							adjustment.ReviewState = reviewstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())

					mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
						proposal.ID, adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadProposalsTest(ReviewableReadProposalsTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/approval-rulesets/ruleset1/proposals",
			Setup:       Setup,
		})

		It("outputs approval rules", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			version := items[0].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("GET /approval-rulesets/:id/proposals/:version_id", func() {
		var mockVersion dbmodels.ApprovalRulesetVersion
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func(approved bool) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				var ruleset dbmodels.ApprovalRuleset

				if approved {
					ruleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					mockVersion = *ruleset.Version
				} else {
					ruleset, err = dbmodels.CreateMockApprovalRuleset(ctx.Db, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					mockVersion, err = dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, ruleset, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, mockVersion, 1,
						func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
							adjustment.ReviewState = reviewstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())

					ruleset.Version = &mockVersion
					ruleset.Version.Adjustment = &adjustment
				}

				app, err := dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
					ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadProposalTest(ReviewableReadProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockVersion.ID) },
			Setup:       Setup,

			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
		})

		It("outputs application bindings", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})
			Expect(appBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeEmpty())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("version", Not(BeEmpty())))
			version := body["version"].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id/proposals/:version_id", func() {
		var mockRuleset dbmodels.ApprovalRuleset
		var mockVersion dbmodels.ApprovalRulesetVersion
		var mockProposal dbmodels.ApprovalRulesetVersion

		Setup := func(reviewState reviewstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				mockRuleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				mockVersion = *mockRuleset.Version

				mockProposal, err = dbmodels.CreateMockApprovalRulesetVersion(ctx.Db, mockRuleset, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(ctx.Db, mockProposal, 1,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ReviewState = reviewState
					})
				Expect(err).ToNot(HaveOccurred())
				mockProposal.Adjustment = &adjustment

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

		includedTestCtx := IncludeReviewableUpdateProposalTest(ReviewableUpdateProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetPath: func(approved bool) string {
				if approved {
					return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockVersion.ID)
				} else {
					return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockProposal.ID)
				}
			},
			Setup: Setup,
			Input: gin.H{
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
			GetPrimaryKey: func(resource interface{}) interface{} {
				return resource.(*dbmodels.ApprovalRuleset).ID
			},
			PrimaryKeyJSONFieldName: "id",
			PrimaryKeyInitialValue:  "ruleset1",
			GetResourceVersionAndLatestAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				version, err := dbmodels.FindApprovalRulesetVersionByID(ctx.Db, ctx.Org.ID, mockRuleset.ID, mockProposal.ID)
				Expect(err).ToNot(HaveOccurred())

				dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRulesetVersion{&version})
				Expect(err).ToNot(HaveOccurred())

				return &version, version.Adjustment
			},
			VersionedFieldJSONFieldName: "display_name",
			VersionedFieldUpdatedValue:  "Ruleset 2",
		})

		It("outputs the updated approval rules", func() {
			Setup(reviewstate.Draft)
			body := includedTestCtx.MakeRequest(false, "", 200)

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			Expect(rules[0]).To(HaveKeyWithValue("type", "schedule"))
			Expect(rules[0]).To(HaveKeyWithValue("begin_time", "1:00"))
			Expect(rules[0]).To(HaveKeyWithValue("end_time", "2:00"))
		})

		It("creates new ApprovalRule objects rather than modifying existing ones in-place", func() {
			Setup(reviewstate.Draft)
			includedTestCtx.MakeRequest(false, "", 200)

			var count int64
			err = ctx.Db.Model(dbmodels.ScheduleApprovalRule{}).Count(&count).Error
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(BeNumerically("==", 3))
		})
	})
})

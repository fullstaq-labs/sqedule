package controllers

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/proposalstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"
)

var _ = Describe("approval-ruleset API", func() {
	var ctx HTTPTestContext
	var err error

	Describe("POST /approval-rulesets", func() {
		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

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
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
			AssertBaseResourceValid: func(resource interface{}) {
				ruleset := resource.(*dbmodels.ApprovalRuleset)
				Expect(ruleset.ID).To(Equal("ruleset1"))
			},
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("display_name", "Ruleset 1"))
			},
			AssertAdjustmentValid: func(adjustment interface{}) {
				a := adjustment.(*dbmodels.ApprovalRulesetAdjustment)
				Expect(a.DisplayName).To(Equal("Ruleset 1"))
			},
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
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				ruleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
			return ruleset
		}

		SetupAssociations := func(ruleset dbmodels.ApprovalRuleset) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app1, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org,
					func(app *dbmodels.Application) {
						app.ID = "app1"
					},
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.DisplayName = "App 1"
					})
				Expect(err).ToNot(HaveOccurred())

				app2, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org,
					func(app *dbmodels.Application) {
						app.ID = "app2"
					},
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.DisplayName = "App 2"
					})
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app1, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app2, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				release1, err := dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app1, nil)
				Expect(err).ToNot(HaveOccurred())
				release2, err := dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app2, nil)
				Expect(err).ToNot(HaveOccurred())
				release3, err := dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app2, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, release1, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, release2, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, release3, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableListResourcesTest(ReviewableListResourcesTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return "/v1/approval-rulesets" },
			Setup:       func() { Setup() },
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
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

		It("does not output rules", func() {
			ruleset := Setup()
			_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org, ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).ToNot(HaveOccurred())

			body := includedTestCtx.MakeRequest()
			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))

			items := body["items"].([]interface{})
			rulesetJSON := items[0].(map[string]interface{})
			Expect(rulesetJSON).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := rulesetJSON["latest_approved_version"].(map[string]interface{})
			Expect(version).ToNot(HaveKey("approval_rules"))
		})
	})

	Describe("GET /approval-rulesets/:id", func() {
		var mockRelease dbmodels.Release
		var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, mockRelease, ruleset,
					*ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
					ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadResourceTest(ReviewableReadResourceTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return "/v1/approval-rulesets/ruleset1" },
			Setup:       Setup,
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
		})

		It("outputs application bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})

			version := appBinding["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeNil())))
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
			Expect(releaseBinding).To(HaveKeyWithValue("release", Not(BeNil())))

			release := releaseBinding["release"].(map[string]interface{})
			Expect(release).To(HaveKeyWithValue("id", BeNumerically("==", mockRelease.ID)))
			Expect(release).To(HaveKeyWithValue("application", Not(BeNil())))

			app := release["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))
			version := body["latest_approved_version"].(map[string]interface{})

			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			rule := rules[0].(map[string]interface{})
			Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
			Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
		})
	})

	Describe("PATCH /approval-rulesets/:id", func() {
		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("upon patching unversioned data", func() {
			var mockRelease dbmodels.Release
			var mockScheduleApprovalRule dbmodels.ScheduleApprovalRule

			Setup := func() {
				_, err = dbmodels.CreateMockApprovalRuleset(ctx.Db, ctx.Org, "ruleset1", nil)
			}

			SetupWithAssocations := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
					Expect(err).ToNot(HaveOccurred())

					mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, mockRelease, ruleset,
						*ruleset.Version, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
						ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					return nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			includedTestCtx := IncludeReviewableUpdateUnversionedDataTest(ReviewableUpdateUnversionedDataTestOptions{
				HTTPTestCtx:      &ctx,
				Path:             "/v1/approval-rulesets/ruleset1",
				Setup:            Setup,
				UnversionedInput: gin.H{"id": "ruleset2"},
				ResourceType:     reflect.TypeOf(dbmodels.ApprovalRuleset{}),
				AssertBaseJSONValid: func(resource map[string]interface{}) {
					Expect(resource).To(HaveKeyWithValue("id", "ruleset2"))
				},
				AssertBaseResourceValid: func(resource interface{}) {
					ruleset := resource.(*dbmodels.ApprovalRuleset)
					Expect(ruleset.ID).To(Equal("ruleset2"))
				},
			})

			It("outputs application bindings", func() {
				SetupWithAssocations()
				body := includedTestCtx.MakeRequest(200)

				Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
				appBindings := body["application_approval_ruleset_bindings"].([]interface{})
				appBinding := appBindings[0].(map[string]interface{})

				version := appBinding["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

				Expect(appBinding).To(HaveKeyWithValue("application", Not(BeNil())))
				app := appBinding["application"].(map[string]interface{})
				Expect(app).To(HaveKeyWithValue("id", "app1"))
			})

			It("outputs release bindings", func() {
				SetupWithAssocations()
				body := includedTestCtx.MakeRequest(200)

				Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
				version := body["version"].(map[string]interface{})

				Expect(version["release_approval_ruleset_bindings"]).To(HaveLen(1))
				releaseBindings := version["release_approval_ruleset_bindings"].([]interface{})
				releaseBinding := releaseBindings[0].(map[string]interface{})
				Expect(releaseBinding).To(HaveKeyWithValue("mode", "enforcing"))
				Expect(releaseBinding).To(HaveKeyWithValue("release", Not(BeNil())))

				release := releaseBinding["release"].(map[string]interface{})
				Expect(release).To(HaveKeyWithValue("id", BeNumerically("==", mockRelease.ID)))
				Expect(release).To(HaveKeyWithValue("application", Not(BeNil())))

				app := release["application"].(map[string]interface{})
				Expect(app).To(HaveKeyWithValue("id", "app1"))
			})

			It("outputs approval rules", func() {
				SetupWithAssocations()
				body := includedTestCtx.MakeRequest(200)

				Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
				version := body["version"].(map[string]interface{})

				Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
				rules := version["approval_rules"].([]interface{})
				rule := rules[0].(map[string]interface{})
				Expect(rule).To(HaveKeyWithValue("id", BeNumerically("==", mockScheduleApprovalRule.ID)))
				Expect(rule).To(HaveKeyWithValue("begin_time", mockScheduleApprovalRule.BeginTime.String))
			})
		})

		Describe("upon patching versioned data", func() {
			var mockRuleset dbmodels.ApprovalRuleset

			Setup := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					mockRuleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, mockRuleset, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
						mockRuleset.Version.ID, *mockRuleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
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

		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRuleset(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					// Create a ruleset with 3 versions
					rulesetVersion1, err := dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(tx, rulesetVersion1, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					// We deliberately create version 3 out of order so that we test
					// whether the versions are outputted in order.

					rulesetVersion3, err := dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, lib.NewUint32Ptr(3), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(tx, rulesetVersion3, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					rulesetVersion2, err := dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, lib.NewUint32Ptr(2), nil)
					Expect(err).ToNot(HaveOccurred())
					rulesetVersion2Adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, rulesetVersion2, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					ruleset.Version = &rulesetVersion3
					ruleset.Version.Adjustment = &rulesetVersion2Adjustment

					app, err := dbmodels.CreateMockApplication(tx, ctx.Org, nil)
					Expect(err).ToNot(HaveOccurred())

					release1, err := dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
					Expect(err).ToNot(HaveOccurred())
					release2, err := dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, release1,
						ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, release2,
						ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
						ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					proposal, err := dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(tx, proposal, 1,
						func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
							adjustment.ProposalState = proposalstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableListVersionsTest(ReviewableListVersionsTestOptions{
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

			Expect(body).To(HaveKeyWithValue("items", Not(BeNil())))
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
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockRelease, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, mockRelease,
					ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
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

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
		})

		It("outputs application bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})

			version := appBinding["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeNil())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs release bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})

			Expect(version["release_approval_ruleset_bindings"]).To(HaveLen(1))
			releaseBindings := version["release_approval_ruleset_bindings"].([]interface{})
			releaseBinding := releaseBindings[0].(map[string]interface{})
			Expect(releaseBinding).To(HaveKeyWithValue("mode", "enforcing"))
			Expect(releaseBinding).To(HaveKeyWithValue("release", Not(BeNil())))

			release := releaseBinding["release"].(map[string]interface{})
			Expect(release).To(HaveKeyWithValue("id", BeNumerically("==", mockRelease.ID)))
			Expect(release).To(HaveKeyWithValue("application", Not(BeNil())))

			app := release["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
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

		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRuleset(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					version, err := dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApprovalRulesetAdjustment(tx, version, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					proposal, err := dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, nil, nil)
					Expect(err).ToNot(HaveOccurred())
					adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, proposal, 1,
						func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
							adjustment.ProposalState = proposalstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())

					mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
						proposal.ID, adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableListProposalsTest(ReviewableListProposalsTestOptions{
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

		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				var ruleset dbmodels.ApprovalRuleset

				if versionIsApproved {
					ruleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					mockVersion = *ruleset.Version
				} else {
					ruleset, err = dbmodels.CreateMockApprovalRuleset(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					mockVersion, err = dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockVersion, 1,
						func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
							adjustment.ProposalState = proposalstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())

					ruleset.Version = &mockVersion
					ruleset.Version.Adjustment = &adjustment
				}

				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				mockScheduleApprovalRule, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
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

			ResourceTypeNameInResponse: "approval ruleset proposal",

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
		})

		It("outputs application bindings", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest(200)

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})

			version := appBinding["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeNil())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest(200)

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
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
		var mockProposal1, mockProposal2 dbmodels.ApprovalRulesetVersion

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Setup := func(hasApprovedVersion bool, proposal1State proposalstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				if hasApprovedVersion {
					mockRuleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					mockVersion = *mockRuleset.Version

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
						mockRuleset.Version.ID, *mockRuleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
						mockRuleset.Version.ID, *mockRuleset.Version.Adjustment, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					mockRuleset, err = dbmodels.CreateMockApprovalRuleset(tx, ctx.Org, "ruleset1", nil)
				}
				Expect(err).ToNot(HaveOccurred())

				mockProposal1, err = dbmodels.CreateMockApprovalRulesetVersion(tx, mockRuleset, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1Adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockProposal1, 1,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ProposalState = proposal1State
					})
				Expect(err).ToNot(HaveOccurred())
				mockProposal1.Adjustment = &proposal1Adjustment

				mockProposal2, err = dbmodels.CreateMockApprovalRulesetVersion(tx, mockRuleset, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal2Adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockProposal2, 1,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ProposalState = proposalstate.Reviewing
					})
				Expect(err).ToNot(HaveOccurred())
				mockProposal2.Adjustment = &proposal2Adjustment

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableUpdateProposalTest(ReviewableUpdateProposalTestOptions{
			HTTPTestCtx:            &ctx,
			GetProposalPath:        func() string { return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockProposal1.ID) },
			GetApprovedVersionPath: func() string { return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockVersion.ID) },
			Setup:                  Setup,
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
			AutoApproveInput:           gin.H{"display_name": "Ruleset 2"},
			AdjustmentType:             reflect.TypeOf(dbmodels.ApprovalRulesetAdjustment{}),
			ResourceTypeNameInResponse: "approval ruleset proposal",
			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
			GetResourceVersionAndLatestAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				version, err := dbmodels.FindApprovalRulesetVersionByID(ctx.Db, ctx.Org.ID, mockRuleset.ID, mockProposal1.ID)
				Expect(err).ToNot(HaveOccurred())

				err = dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRulesetVersion{&version})
				Expect(err).ToNot(HaveOccurred())

				return &version, version.Adjustment
			},
			VersionedFieldJSONFieldName: "display_name",
			VersionedFieldUpdatedValue:  "Ruleset 2",
			GetSecondProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApprovalRulesetVersion
				tx := ctx.Db.Where("id = ?", mockProposal2.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRulesetVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
		})

		It("outputs application bindings", func() {
			Setup(true, proposalstate.Draft)

			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, mockRuleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			body := includedTestCtx.MakeRequest(false, false, "", 200)

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})

			version := appBinding["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeNil())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs the updated approval rules", func() {
			Setup(true, proposalstate.Draft)
			body := includedTestCtx.MakeRequest(false, false, "", 200)

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			Expect(rules[0]).To(HaveKeyWithValue("type", "schedule"))
			Expect(rules[0]).To(HaveKeyWithValue("begin_time", "1:00"))
			Expect(rules[0]).To(HaveKeyWithValue("end_time", "2:00"))
		})

		It("creates new ApprovalRule objects rather than modifying existing ones in-place", func() {
			Setup(true, proposalstate.Draft)
			includedTestCtx.MakeRequest(false, false, "", 200)

			var count int64
			err = ctx.Db.Model(dbmodels.ScheduleApprovalRule{}).Count(&count).Error
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(BeNumerically("==", 3))
		})

		It("copies the previous adjustment's approval rules if no new approval rules are given", func() {
			Setup(true, proposalstate.Draft)
			rule, err := dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.Db, ctx.Org,
				mockProposal1.ID, *mockProposal1.Adjustment, nil)
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("PATCH", fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d",
				mockProposal1.ID), gin.H{})
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.Recorder.Code).To(BeNumerically("==", 200))
			body, err := ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			Expect(rules[0]).To(HaveKeyWithValue("type", "schedule"))
			Expect(rules[0]).To(HaveKeyWithValue("begin_time", rule.BeginTime.String))
			Expect(rules[0]).To(HaveKeyWithValue("end_time", rule.EndTime.String))
		})
	})

	Describe("PUT /approval-rulesets/:id/proposals/:version_id/state", func() {
		var mockVersion dbmodels.ApprovalRulesetVersion
		var mockProposal1, mockProposal2 dbmodels.ApprovalRulesetVersion

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Setup := func(proposalState proposalstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())
				mockVersion = *ruleset.Version

				mockProposal1, err = dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1Adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockProposal1, 1,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ProposalState = proposalState
					})
				Expect(err).ToNot(HaveOccurred())
				mockProposal1.Adjustment = &proposal1Adjustment

				_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org,
					mockProposal1.ID, proposal1Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				mockProposal2, err = dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal2Adjustment, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockProposal2, 1,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ProposalState = proposalState
					})
				Expect(err).ToNot(HaveOccurred())
				mockProposal2.Adjustment = &proposal2Adjustment

				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReviewProposalTest(ReviewableReviewProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetProposalPath: func() string {
				return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d/state", mockProposal1.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d/state", mockVersion.ID)
			},
			Setup:                      Setup,
			ResourceTypeNameInResponse: "approval ruleset proposal",
			GetFirstProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApprovalRulesetVersion
				tx := ctx.Db.Where("id = ?", mockProposal1.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRulesetVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
			GetSecondProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApprovalRulesetVersion
				tx := ctx.Db.Where("id = ?", mockProposal2.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApprovalRulesetVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApprovalRulesetVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "ruleset1"))
			},
			VersionedFieldJSONFieldName: "display_name",
			VersionedFieldInitialValue:  "Ruleset",
		})

		It("outputs application bindings", func() {
			Setup(proposalstate.Reviewing)
			body := includedTestCtx.MakeRequest(false, "approved", 200)

			Expect(body).To(HaveKeyWithValue("application_approval_ruleset_bindings", HaveLen(1)))
			appBindings := body["application_approval_ruleset_bindings"].([]interface{})
			appBinding := appBindings[0].(map[string]interface{})

			version := appBinding["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

			Expect(appBinding).To(HaveKeyWithValue("application", Not(BeNil())))
			app := appBinding["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
		})

		It("outputs approval rules", func() {
			Setup(proposalstate.Reviewing)
			body := includedTestCtx.MakeRequest(false, "approved", 200)

			Expect(body).To(HaveKeyWithValue("version", Not(BeNil())))
			version := body["version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("approval_rules", HaveLen(1)))
			rules := version["approval_rules"].([]interface{})
			Expect(rules[0]).To(HaveKeyWithValue("type", "schedule"))
		})

		It("creates copies of ApprovalRule objects", func() {
			Setup(proposalstate.Reviewing)
			includedTestCtx.MakeRequest(false, "approved", 200)

			var count int64
			err = ctx.Db.Model(dbmodels.ScheduleApprovalRule{}).Count(&count).Error
			Expect(err).ToNot(HaveOccurred())
			Expect(count).To(BeNumerically("==", 2))
		})
	})

	Describe("DELETE /approval-rulesets/:id/proposals/:version_id", func() {
		var mockVersion dbmodels.ApprovalRulesetVersion
		var mockProposal dbmodels.ApprovalRulesetVersion

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())
				mockVersion = *ruleset.Version

				// Create a proposal with 2 adjustments

				mockProposal, err = dbmodels.CreateMockApprovalRulesetVersion(tx, ruleset, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				// Adjustment 1

				adjustment1, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockProposal, 1,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ProposalState = proposalstate.Draft
					})
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockCreationAuditRecord(tx, ctx.Org,
					func(record *dbmodels.CreationAuditRecord) {
						record.ApprovalRulesetVersionID = &mockProposal.ID
						record.ApprovalRulesetAdjustmentNumber = &adjustment1.AdjustmentNumber
					})
				Expect(err).ToNot(HaveOccurred())

				// Adjustment 2

				adjustment2, err := dbmodels.CreateMockApprovalRulesetAdjustment(tx, mockProposal, 2,
					func(adjustment *dbmodels.ApprovalRulesetAdjustment) {
						adjustment.ProposalState = proposalstate.Draft
					})
				Expect(err).ToNot(HaveOccurred())
				mockProposal.Adjustment = &adjustment2

				_, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org, mockProposal.ID, adjustment2, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockCreationAuditRecord(tx, ctx.Org,
					func(record *dbmodels.CreationAuditRecord) {
						record.ApprovalRulesetVersionID = &mockProposal.ID
						record.ApprovalRulesetAdjustmentNumber = &adjustment2.AdjustmentNumber
					})
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableDeleteProposalTest(ReviewableDeleteProposalTestOptions{
			HTTPTestCtx:                &ctx,
			GetProposalPath:            func() string { return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockProposal.ID) },
			GetApprovedVersionPath:     func() string { return fmt.Sprintf("/v1/approval-rulesets/ruleset1/proposals/%d", mockVersion.ID) },
			Setup:                      Setup,
			ResourceTypeNameInResponse: "approval ruleset proposal",
			CountProposals: func() uint {
				var count int64
				err = ctx.Db.Model(dbmodels.ApprovalRulesetVersion{}).Where("version_number IS NULL").Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				return uint(count)
			},
			CountProposalAdjustments: func() uint {
				var count int64
				err = ctx.Db.Model(dbmodels.ApprovalRulesetAdjustment{}).Where("approval_ruleset_version_id = ?", mockProposal.ID).Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				return uint(count)
			},
		})
	})
})

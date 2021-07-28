package controllers

import (
	"fmt"
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("application API", func() {
	var ctx HTTPTestContext
	var err error

	Describe("POST /applications", func() {
		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		includedTestCtx := IncludeReviewableCreateTest(ReviewableCreateTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/applications/app1",
			UnversionedInput: gin.H{
				"id": "app1",
			},
			VersionedInput: gin.H{
				"display_name": "New App",
			},
			ResourceType:   reflect.TypeOf(dbmodels.Application{}),
			AdjustmentType: reflect.TypeOf(dbmodels.ApplicationAdjustment{}),
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "app1"))
			},
			AssertBaseResourceValid: func(resource interface{}) {
				binding := resource.(*dbmodels.Application)
				Expect(binding.ID).To(Equal("app1"))
			},
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("display_name", "New App"))
			},
			AssertAdjustmentValid: func(adjustment interface{}) {
				a := adjustment.(*dbmodels.ApplicationAdjustment)
				Expect(a.DisplayName).To(Equal("New App"))
			},
		})

		It("outputs no approval ruleset bindings", func() {
			body := includedTestCtx.MakeRequest("", 201)
			Expect(body).ToNot(HaveKey("approval_ruleset_bindings"))
		})
	})

	Describe("GET /applications", func() {
		var app dbmodels.Application

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app,
					ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableListResourcesTest(ReviewableListResourcesTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return "/v1/applications" },
			Setup:       Setup,
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", app.ID))
			},
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("display_name", app.Version.Adjustment.DisplayName))
				Expect(version).To(HaveKeyWithValue("enabled", app.Version.Adjustment.IsEnabled()))
			},
		})

		It("outputs no approval ruleset bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).ToNot(HaveKey("approval_ruleset_bindings"))
		})
	})

	Describe("GET /applications/:id", func() {
		var app dbmodels.Application

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadResourceTest(ReviewableReadResourceTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return "/v1/applications/app1" },
			Setup:       Setup,
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", app.ID))
			},
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("display_name", app.Version.Adjustment.DisplayName))
				Expect(version).To(HaveKeyWithValue("enabled", app.Version.Adjustment.IsEnabled()))
			},
		})

		It("outputs approval ruleset bindings", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("approval_ruleset_bindings", HaveLen(1)))
			bindings := body["approval_ruleset_bindings"].([]interface{})
			binding := bindings[0].(map[string]interface{})

			version := binding["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

			Expect(binding).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
			ruleset := binding["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version = ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
		})
	})

	Describe("PATCH /applications/:id", func() {
		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("upon patching unversioned data", func() {
			Setup := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
					Expect(err).ToNot(HaveOccurred())

					return nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			includedTestCtx := IncludeReviewableUpdateUnversionedDataTest(ReviewableUpdateUnversionedDataTestOptions{
				HTTPTestCtx:      &ctx,
				Path:             "/v1/applications/app1",
				Setup:            Setup,
				UnversionedInput: gin.H{},
				ResourceType:     reflect.TypeOf(dbmodels.Application{}),
			})

			It("outputs approval ruleset bindings", func() {
				Setup()
				body := includedTestCtx.MakeRequest(200)

				Expect(body).To(HaveKeyWithValue("approval_ruleset_bindings", HaveLen(1)))
				bindings := body["approval_ruleset_bindings"].([]interface{})
				binding := bindings[0].(map[string]interface{})

				version := binding["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

				Expect(binding).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				ruleset := binding["approval_ruleset"].(map[string]interface{})
				Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
				Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

				version = ruleset["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
			})
		})

		Describe("upon patching versioned data", func() {
			var app dbmodels.Application

			Setup := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
					Expect(err).ToNot(HaveOccurred())

					return nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			includedTestCtx := IncludeReviewableUpdateVersionedDataTest(ReviewableUpdateVersionedDataTestOptions{
				HTTPTestCtx:    &ctx,
				Path:           "/v1/applications/app1",
				Setup:          Setup,
				VersionedInput: gin.H{"display_name": "Changed Name"},
				AdjustmentType: reflect.TypeOf(dbmodels.ApplicationAdjustment{}),
				GetLatestResourceVersionAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
					err := dbmodels.LoadApplicationsLatestVersionsAndAdjustments(ctx.Db, ctx.Org.ID,
						[]*dbmodels.Application{&app})
					Expect(err).ToNot(HaveOccurred())
					return app.Version, app.Version.Adjustment
				},
				VersionedFieldJSONFieldName: "display_name",
				VersionedFieldUpdatedValue:  "Changed Name",
			})

			It("outputs approval ruleset bindings", func() {
				Setup()
				body := includedTestCtx.MakeRequest("", 200)

				Expect(body).To(HaveKeyWithValue("approval_ruleset_bindings", HaveLen(1)))
				bindings := body["approval_ruleset_bindings"].([]interface{})
				binding := bindings[0].(map[string]interface{})

				version := binding["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))

				Expect(binding).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				ruleset := binding["approval_ruleset"].(map[string]interface{})
				Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
				Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

				version = ruleset["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
			})
		})
	})
})

package controllers

import (
	"fmt"
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("application-approval-ruleset-binding API", func() {
	var ctx HTTPTestContext
	var err error

	Describe("POST /applications/:application_id/approval-ruleset-bindings/:ruleset_id", func() {
		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				_, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		})

		includedTestCtx := IncludeReviewableCreateTest(ReviewableCreateTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/applications/app1/approval-ruleset-bindings/ruleset1",
			UnversionedInput: gin.H{
				"id": "ruleset1",
			},
			VersionedInput: gin.H{
				"mode": "permissive",
			},
			ResourceType:   reflect.TypeOf(dbmodels.ApplicationApprovalRulesetBinding{}),
			AdjustmentType: reflect.TypeOf(dbmodels.ApplicationApprovalRulesetBindingAdjustment{}),
			AssertBaseJSONValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				Expect(resource["approval_ruleset"]).To(HaveKeyWithValue("id", "ruleset1"))
			},
			AssertBaseResourceValid: func(resource interface{}) {
				binding := resource.(*dbmodels.ApplicationApprovalRulesetBinding)
				Expect(binding.ApplicationID).To(Equal("app1"))
				Expect(binding.ApprovalRulesetID).To(Equal("ruleset1"))
			},
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "permissive"))
			},
			AssertAdjustmentValid: func(adjustment interface{}) {
				a := adjustment.(*dbmodels.ApplicationApprovalRulesetBindingAdjustment)
				Expect(a.Mode).To(Equal(approvalrulesetbindingmode.Permissive))
			},
		})

		It("output no application", func() {
			body := includedTestCtx.MakeRequest("", 201)
			Expect(body).ToNot(HaveKey("application"))
		})

		It("outputs approval ruleset", func() {
			body := includedTestCtx.MakeRequest("", 201)
			Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := body["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("GET /applications/:application_id/approval-ruleset-bindings", func() {
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

		includedTestCtx := IncludeReviewableListResourcesTest(ReviewableListResourcesTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return fmt.Sprintf("/v1/applications/%s/approval-ruleset-bindings", app.ID) },
			Setup:       Setup,
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))
			},
		})

		It("outputs no applications", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).ToNot(HaveKey("application"))
		})

		It("outputs approval rulesets", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
		})
	})

	Describe("GET /applications/:application_id/approval-ruleset-bindings/:ruleset_id", func() {
		var app dbmodels.Application
		var binding dbmodels.ApplicationApprovalRulesetBinding

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				binding, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadResourceTest(ReviewableReadResourceTestOptions{
			HTTPTestCtx: &ctx,
			GetPath: func() string {
				return fmt.Sprintf("/v1/applications/%s/approval-ruleset-bindings/%s", app.ID, binding.ApprovalRulesetID)
			},
			Setup: Setup,
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))
			},
		})

		It("outputs no application", func() {
			Setup()
			body := includedTestCtx.MakeRequest()
			Expect(body).ToNot(HaveKey("application"))
		})

		It("outputs approval ruleset", func() {
			Setup()
			body := includedTestCtx.MakeRequest()
			Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := body["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("PATCH /applications/:application_id/approval-rulesets-bindings/:ruleset_id", func() {
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
				Path:             "/v1/applications/app1/approval-ruleset-bindings/ruleset1",
				Setup:            Setup,
				UnversionedInput: gin.H{},
				ResourceType:     reflect.TypeOf(dbmodels.ApplicationApprovalRulesetBinding{}),
			})

			It("outputs no application", func() {
				Setup()
				body := includedTestCtx.MakeRequest(200)
				Expect(body).ToNot(HaveKey("application"))
			})

			It("outputs approval ruleset", func() {
				Setup()
				body := includedTestCtx.MakeRequest(200)
				Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

				ruleset := body["approval_ruleset"].(map[string]interface{})
				Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
				Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

				version := ruleset["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
			})
		})

		Describe("upon patching versioned data", func() {
			var binding dbmodels.ApplicationApprovalRulesetBinding

			Setup := func() {
				err = ctx.Db.Transaction(func(tx *gorm.DB) error {
					app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
					Expect(err).ToNot(HaveOccurred())

					binding, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
					Expect(err).ToNot(HaveOccurred())

					return nil
				})
				Expect(err).ToNot(HaveOccurred())
			}

			includedTestCtx := IncludeReviewableUpdateVersionedDataTest(ReviewableUpdateVersionedDataTestOptions{
				HTTPTestCtx:    &ctx,
				Path:           "/v1/applications/app1/approval-ruleset-bindings/ruleset1",
				Setup:          Setup,
				VersionedInput: gin.H{"mode": "permissive"},
				AdjustmentType: reflect.TypeOf(dbmodels.ApplicationApprovalRulesetBindingAdjustment{}),
				GetLatestResourceVersionAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
					err := dbmodels.LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(ctx.Db, ctx.Org.ID,
						[]*dbmodels.ApplicationApprovalRulesetBinding{&binding})
					Expect(err).ToNot(HaveOccurred())
					return binding.Version, binding.Version.Adjustment
				},
				VersionedFieldJSONFieldName: "mode",
				VersionedFieldUpdatedValue:  "permissive",
			})

			It("outputs no application", func() {
				Setup()
				body := includedTestCtx.MakeRequest("", 200)
				Expect(body).ToNot(HaveKey("application"))
			})

			It("outputs approval ruleset", func() {
				Setup()
				body := includedTestCtx.MakeRequest("", 200)
				Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

				ruleset := body["approval_ruleset"].(map[string]interface{})
				Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
				Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

				version := ruleset["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
			})
		})
	})

	Describe("GET /applications/:application_id/approval-ruleset-bindings/:ruleset_id/versions", func() {
		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				binding, err := dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					// Create a binding with 3 versions
					version1, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.Org, version1, nil)
					Expect(err).ToNot(HaveOccurred())

					// We deliberately create version 3 out of order so that we test
					// whether the versions are outputted in order.

					version3, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(3), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.Org, version3, nil)
					Expect(err).ToNot(HaveOccurred())

					version2, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(2), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.Org, version2, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.Org, version,
						func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
							adjustment.ReviewState = reviewstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableListVersionsTest(ReviewableListVersionsTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/applications/app1/approval-ruleset-bindings/ruleset1/versions",
			Setup:       Setup,
		})
	})

	Describe("GET /applications/:application_id/approval-ruleset-bindings/:ruleset_id/versions/:version_number", func() {
		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
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

		includedTestCtx := IncludeReviewableReadVersionTest(ReviewableReadVersionTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/applications/app1/approval-ruleset-bindings/ruleset1/versions/1",
			Setup:       Setup,

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				Expect(resource["approval_ruleset"]).To(HaveKeyWithValue("id", "ruleset1"))
			},
		})

		It("outputs no application", func() {
			Setup()
			body := includedTestCtx.MakeRequest()
			Expect(body).ToNot(HaveKey("application"))
		})

		It("outputs approval ruleset", func() {
			Setup()
			body := includedTestCtx.MakeRequest()
			Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := body["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})
})

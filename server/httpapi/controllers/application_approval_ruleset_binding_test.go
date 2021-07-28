package controllers

import (
	"fmt"
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/reviewstate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("application-approval-ruleset-binding API", func() {
	var ctx HTTPTestContext
	var err error

	Describe("POST /application-approval-ruleset-bindings/:application_id/:ruleset_id", func() {
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
			Path:        "/v1/application-approval-ruleset-bindings/app1/ruleset1",
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

		It("output application", func() {
			body := includedTestCtx.MakeRequest("", 201)
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
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

	Describe("GET /applications-approval-ruleset-bindings", func() {
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
			GetPath:     func() string { return "/v1/application-approval-ruleset-bindings" },
			Setup:       Setup,
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))
			},
		})

		It("outputs applications", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).To(HaveKeyWithValue("application", Not(BeNil())))

			app := ruleset["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
		})

		It("outputs approval rulesets", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			binding := items[0].(map[string]interface{})

			ruleset := binding["approval_ruleset"].(map[string]interface{})
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
			binding := items[0].(map[string]interface{})

			Expect(binding).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := binding["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("GET /application-approval-ruleset-bindings/:application_id/:ruleset_id", func() {
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
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/%s/%s", app.ID, binding.ApprovalRulesetID)
			},
			Setup: Setup,
			AssertVersionJSONValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))
			},
		})

		It("outputs application", func() {
			Setup()
			body := includedTestCtx.MakeRequest()
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
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

	Describe("PATCH /application-approval-ruleset-bindings/:application_id/:ruleset_id", func() {
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
				Path:             "/v1/application-approval-ruleset-bindings/app1/ruleset1",
				Setup:            Setup,
				UnversionedInput: gin.H{},
				ResourceType:     reflect.TypeOf(dbmodels.ApplicationApprovalRulesetBinding{}),
			})

			It("outputs application", func() {
				Setup()
				body := includedTestCtx.MakeRequest(200)
				Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

				app := body["application"].(map[string]interface{})
				Expect(app).To(HaveKeyWithValue("id", "app1"))
				Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

				version := app["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
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
				Path:           "/v1/application-approval-ruleset-bindings/app1/ruleset1",
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

			It("outputs application", func() {
				Setup()
				body := includedTestCtx.MakeRequest("", 200)
				Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

				app := body["application"].(map[string]interface{})
				Expect(app).To(HaveKeyWithValue("id", "app1"))
				Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

				version := app["latest_approved_version"].(map[string]interface{})
				Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
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

	Describe("GET /application-approval-ruleset-bindings/:application_id/:ruleset_id/versions", func() {
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
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version1, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					// We deliberately create version 3 out of order so that we test
					// whether the versions are outputted in order.

					version3, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(3), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version3, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					version2, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(2), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version2, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version, 1,
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
			Path:        "/v1/application-approval-ruleset-bindings/app1/ruleset1/versions",
			Setup:       Setup,
		})
	})

	Describe("GET /application-approval-ruleset-bindings/:application_id/:ruleset_id/versions/:version_number", func() {
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
			Path:        "/v1/application-approval-ruleset-bindings/app1/ruleset1/versions/1",
			Setup:       Setup,

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				Expect(resource["approval_ruleset"]).To(HaveKeyWithValue("id", "ruleset1"))
			},
		})

		It("outputs application", func() {
			Setup()
			body := includedTestCtx.MakeRequest()
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
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

	Describe("GET /application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals", func() {
		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				binding, err := dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					version, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err := dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version, 1,
						func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
							adjustment.ReviewState = reviewstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableListProposalsTest(ReviewableListProposalsTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals",
			Setup:       Setup,
		})
	})

	Describe("GET /application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id", func() {
		var version dbmodels.ApplicationApprovalRulesetBindingVersion

		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				binding, err := dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					version, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, version, 1,
						func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
							adjustment.ReviewState = reviewstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadProposalTest(ReviewableReadProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d", version.ID)
			},
			Setup: Setup,

			ResourceTypeNameInResponse: "application approval ruleset binding proposal",

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				Expect(resource["approval_ruleset"]).To(HaveKeyWithValue("id", "ruleset1"))
			},
		})

		It("outputs application", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest(200)
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
		})

		It("outputs approval ruleset", func() {
			Setup(false)
			body := includedTestCtx.MakeRequest(200)
			Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := body["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("PATCH /application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id", func() {
		var app dbmodels.Application
		var ruleset dbmodels.ApprovalRuleset
		var proposal1, proposal2, version dbmodels.ApplicationApprovalRulesetBindingVersion

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Setup := func(hasApprovedVersion bool, proposal1ReviewState reviewstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err = dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				var binding dbmodels.ApplicationApprovalRulesetBinding
				if hasApprovedVersion {
					binding, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
					version = *binding.Version
				} else {
					binding, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode(tx, ctx.Org, app, ruleset, nil)
				}
				Expect(err).ToNot(HaveOccurred())

				proposal1, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1Adjustment, err := dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, proposal1, 1,
					func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
						adjustment.ReviewState = proposal1ReviewState
					})
				Expect(err).ToNot(HaveOccurred())
				proposal1.Adjustment = &proposal1Adjustment

				proposal2, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal2Adjustment, err := dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, proposal2, 1,
					func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
						adjustment.ReviewState = reviewstate.Reviewing
					})
				Expect(err).ToNot(HaveOccurred())
				proposal2.Adjustment = &proposal2Adjustment

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableUpdateProposalTest(ReviewableUpdateProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetProposalPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d", proposal1.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d", version.ID)
			},
			Setup:                      Setup,
			Input:                      gin.H{"mode": "permissive"},
			AutoApproveInput:           gin.H{},
			AdjustmentType:             reflect.TypeOf(dbmodels.ApplicationApprovalRulesetBindingAdjustment{}),
			ResourceTypeNameInResponse: "application approval ruleset binding proposal",
			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				Expect(resource["approval_ruleset"]).To(HaveKeyWithValue("id", "ruleset1"))
			},
			GetResourceVersionAndLatestAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				version, err := dbmodels.FindApplicationApprovalRulesetBindingVersionByID(ctx.Db, ctx.Org.ID, app.ID, ruleset.ID, proposal1.ID)
				Expect(err).ToNot(HaveOccurred())

				dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, ctx.Org.ID,
					[]*dbmodels.ApplicationApprovalRulesetBindingVersion{&version})
				Expect(err).ToNot(HaveOccurred())

				return &version, version.Adjustment
			},
			VersionedFieldJSONFieldName: "mode",
			VersionedFieldUpdatedValue:  "permissive",
			GetSecondProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApplicationApprovalRulesetBindingVersion
				tx := ctx.Db.Where("id = ?", proposal2.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, ctx.Org.ID,
					[]*dbmodels.ApplicationApprovalRulesetBindingVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
		})

		It("outputs no application", func() {
			Setup(true, reviewstate.Draft)
			body := includedTestCtx.MakeRequest(false, false, "", 200)
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
		})

		It("outputs approval ruleset", func() {
			Setup(true, reviewstate.Draft)
			body := includedTestCtx.MakeRequest(false, false, "", 200)
			Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := body["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("PUT /application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id/review-state", func() {
		var proposal1, proposal2, version dbmodels.ApplicationApprovalRulesetBindingVersion

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Setup := func(reviewState reviewstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				binding, err := dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				version = *binding.Version
				Expect(err).ToNot(HaveOccurred())

				proposal1, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1Adjustment, err := dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, proposal1, 1,
					func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
						adjustment.ReviewState = reviewState
					})
				Expect(err).ToNot(HaveOccurred())
				proposal1.Adjustment = &proposal1Adjustment

				proposal2, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal2Adjustment, err := dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, proposal2, 1,
					func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
						adjustment.ReviewState = reviewState
					})
				Expect(err).ToNot(HaveOccurred())
				proposal2.Adjustment = &proposal2Adjustment

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReviewProposalTest(ReviewableReviewProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetProposalPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d/review-state", proposal1.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d/review-state", version.ID)
			},
			Setup:                      Setup,
			ResourceTypeNameInResponse: "application approval ruleset binding proposal",
			GetFirstProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApplicationApprovalRulesetBindingVersion
				tx := ctx.Db.Where("id = ?", proposal1.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, ctx.Org.ID,
					[]*dbmodels.ApplicationApprovalRulesetBindingVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
			GetSecondProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApplicationApprovalRulesetBindingVersion
				tx := ctx.Db.Where("id = ?", proposal2.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApplicationApprovalRulesetBindingVersionsLatestAdjustments(ctx.Db, ctx.Org.ID,
					[]*dbmodels.ApplicationApprovalRulesetBindingVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
				Expect(resource["approval_ruleset"]).To(HaveKeyWithValue("id", "ruleset1"))
			},
			VersionedFieldJSONFieldName: "mode",
			VersionedFieldInitialValue:  "enforcing",
		})

		It("outputs application", func() {
			Setup(reviewstate.Reviewing)
			body := includedTestCtx.MakeRequest(false, "approved", 200)
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "App 1"))
		})

		It("outputs approval ruleset", func() {
			Setup(reviewstate.Reviewing)
			body := includedTestCtx.MakeRequest(false, "approved", 200)
			Expect(body).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))

			ruleset := body["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("DELETE /application-approval-ruleset-bindings/:application_id/:ruleset_id/proposals/:version_id", func() {
		var version, proposal dbmodels.ApplicationApprovalRulesetBindingVersion

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				binding, err := dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				version = *binding.Version
				Expect(err).ToNot(HaveOccurred())

				// Create a proposal with 2 adjustments

				proposal, err = dbmodels.CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.Org, app, binding, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				// Adjustment 1

				adjustment1, err := dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, proposal, 1,
					func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
						adjustment.ReviewState = reviewstate.Draft
					})
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockCreationAuditRecord(tx, ctx.Org,
					func(record *dbmodels.CreationAuditRecord) {
						record.ApplicationApprovalRulesetBindingVersionID = &proposal.ID
						record.ApplicationApprovalRulesetBindingAdjustmentNumber = &adjustment1.AdjustmentNumber
					})
				Expect(err).ToNot(HaveOccurred())

				// Adjustment 2

				adjustment2, err := dbmodels.CreateMockApplicationApprovalRulesetBindingAdjustment(tx, proposal, 2,
					func(adjustment *dbmodels.ApplicationApprovalRulesetBindingAdjustment) {
						adjustment.ReviewState = reviewstate.Draft
					})
				Expect(err).ToNot(HaveOccurred())
				proposal.Adjustment = &adjustment2

				_, err = dbmodels.CreateMockCreationAuditRecord(tx, ctx.Org,
					func(record *dbmodels.CreationAuditRecord) {
						record.ApplicationApprovalRulesetBindingVersionID = &proposal.ID
						record.ApplicationApprovalRulesetBindingAdjustmentNumber = &adjustment2.AdjustmentNumber
					})
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableDeleteProposalTest(ReviewableDeleteProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetProposalPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d", proposal.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/application-approval-ruleset-bindings/app1/ruleset1/proposals/%d", version.ID)
			},
			Setup:                      Setup,
			ResourceTypeNameInResponse: "application approval ruleset binding proposal",
			CountProposals: func() uint {
				var count int64
				err = ctx.Db.Model(dbmodels.ApplicationApprovalRulesetBindingVersion{}).Where("version_number IS NULL").Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				return uint(count)
			},
			CountProposalAdjustments: func() uint {
				var count int64
				err = ctx.Db.Model(dbmodels.ApplicationApprovalRulesetBindingAdjustment{}).Where("application_approval_ruleset_binding_version_id = ?", proposal.ID).Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				return uint(count)
			},
		})
	})
})

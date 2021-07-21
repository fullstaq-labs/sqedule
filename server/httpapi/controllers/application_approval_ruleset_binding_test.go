package controllers

import (
	"fmt"
	"reflect"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
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
				Expect(resource).To(HaveKeyWithValue("application", Not(BeNil())))
				Expect(resource["application"]).To(HaveKeyWithValue("id", "app1"))
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

		It("outputs applications", func() {
			body := includedTestCtx.MakeRequest("", 201)
			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))

			app := body["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", "app1"))
			Expect(app).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			version := app["latest_approved_version"].(map[string]interface{})
			Expect(version["display_name"]).To(Equal("App 1"))
		})

		It("outputs approval rulesets", func() {
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

		It("outputs approval rulesets", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).To(HaveKeyWithValue("approval_ruleset", Not(BeNil())))
		})

		It("outputs no applications", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).ToNot(HaveKey("application"))
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

		It("outputs applications", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application", Not(BeNil())))
			appJSON := body["application"].(map[string]interface{})
			Expect(appJSON).To(HaveKeyWithValue("id", app.ID))
			Expect(appJSON).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))
		})
	})
})

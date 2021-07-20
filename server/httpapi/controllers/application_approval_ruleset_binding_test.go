package controllers

import (
	"fmt"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("application-approval-ruleset-binding API", func() {
	var ctx HTTPTestContext
	var err error

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
			AssertVersionValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))
			},
		})

		It("outputs approval rulesets", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("items", HaveLen(1)))
			items := body["items"].([]interface{})
			ruleset := items[0].(map[string]interface{})

			Expect(ruleset).To(HaveKeyWithValue("approval_ruleset", Not(BeEmpty())))
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
			AssertVersionValid: func(version map[string]interface{}) {
				Expect(version).To(HaveKeyWithValue("mode", "enforcing"))
			},
		})

		It("outputs applications", func() {
			Setup()
			body := includedTestCtx.MakeRequest()

			Expect(body).To(HaveKeyWithValue("application", Not(BeEmpty())))
			appJSON := body["application"].(map[string]interface{})
			Expect(appJSON).To(HaveKeyWithValue("id", app.ID))
			Expect(appJSON).To(HaveKeyWithValue("latest_approved_version", Not(BeEmpty())))
		})
	})
})

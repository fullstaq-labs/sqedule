package controllers

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("application API", func() {
	var ctx HTTPTestContext
	var err error

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
			AssertBaseResourceValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", app.ID))
			},
			AssertVersionValid: func(version map[string]interface{}) {
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

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app,
					ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		includedTestCtx := IncludeReviewableReadResourceTest(ReviewableReadResourceTestOptions{
			HTTPTestCtx: &ctx,
			GetPath:     func() string { return "/v1/applications/app1" },
			Setup:       Setup,
			AssertBaseResourceValid: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", app.ID))
			},
			AssertVersionValid: func(version map[string]interface{}) {
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

			Expect(binding).To(HaveKeyWithValue("approval_ruleset", Not(BeEmpty())))
			ruleset := binding["approval_ruleset"].(map[string]interface{})
			Expect(ruleset).To(HaveKeyWithValue("id", "ruleset1"))
			Expect(ruleset).To(HaveKeyWithValue("latest_approved_version", Not(BeEmpty())))

			version = ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version).To(HaveKeyWithValue("version_number", BeNumerically("==", 1)))
		})
	})
})

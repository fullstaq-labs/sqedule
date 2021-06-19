package controllers

import (
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("approval-ruleset API", func() {
	var ctx HTTPTestContext
	var err error

	BeforeEach(func() {
		ctx, err = SetupHTTPTestContext()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("GET /applications", func() {
		var app1, app2 dbmodels.Application
		var body gin.H

		BeforeEach(func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app1, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
					func(app *dbmodels.Application) {
						app.CreatedAt = time.Date(2021, 3, 8, 12, 0, 0, 0, time.Local)
					},
					nil)
				Expect(err).ToNot(HaveOccurred())

				app2, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
					func(app *dbmodels.Application) {
						app.ID = "app2"
						app.CreatedAt = time.Date(2021, 3, 8, 11, 0, 0, 0, time.Local)
					},
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.DisplayName = "App 2"
					})
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app1,
					ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications"), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs all applications", func() {
			Expect(body["items"]).NotTo(BeNil())
			items := body["items"].([]interface{})
			Expect(items).To(HaveLen(2))

			item1 := items[0].(map[string]interface{})
			Expect(item1["id"]).To(Equal(app1.ID))
			Expect(item1["version_number"]).To(Equal(float64(1)))
			Expect(item1["adjustment_number"]).To(Equal(float64(1)))
			Expect(item1["display_name"]).To(Equal(app1.Version.Adjustment.DisplayName))
			Expect(item1["enabled"]).To(Equal(app1.Version.Adjustment.Enabled))
			Expect(item1["created_at"]).ToNot(BeNil())
			Expect(item1["updated_at"]).ToNot(BeNil())

			item2 := items[1].(map[string]interface{})
			Expect(item2["id"]).To(Equal(app2.ID))
			Expect(item2["version_number"]).To(Equal(float64(1)))
			Expect(item2["adjustment_number"]).To(Equal(float64(1)))
			Expect(item2["display_name"]).To(Equal(app2.Version.Adjustment.DisplayName))
			Expect(item2["enabled"]).To(Equal(app2.Version.Adjustment.Enabled))
			Expect(item2["created_at"]).ToNot(BeNil())
			Expect(item2["updated_at"]).ToNot(BeNil())
		})

		It("outputs no approval ruleset bindings", func() {
			Expect(body["items"]).NotTo(BeNil())
			items := body["items"].([]interface{})
			Expect(items).To(HaveLen(2))

			item1 := items[0].(map[string]interface{})
			Expect(item1).ToNot(HaveKey("approval_ruleset_bindings"))

			item2 := items[1].(map[string]interface{})
			Expect(item2).ToNot(HaveKey("approval_ruleset_bindings"))
		})
	})

	Describe("GET /applications/:id", func() {
		var app dbmodels.Application
		var body gin.H

		BeforeEach(func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, app,
					ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s", app.ID), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs the latest version", func() {
			Expect(body["id"]).To(Equal(app.ID))
			Expect(body["version_number"]).To(Equal(float64(1)))
			Expect(body["adjustment_number"]).To(Equal(float64(1)))
			Expect(body["display_name"]).To(Equal(app.Version.Adjustment.DisplayName))
			Expect(body["enabled"]).To(Equal(app.Version.Adjustment.Enabled))
			Expect(body["created_at"]).ToNot(BeNil())
			Expect(body["updated_at"]).ToNot(BeNil())
		})

		It("outputs approval ruleset bindings", func() {
			Expect(body["approval_ruleset_bindings"]).ToNot(BeEmpty())

			bindings := body["approval_ruleset_bindings"].([]interface{})
			Expect(bindings).To(HaveLen(1))

			binding := bindings[0].(map[string]interface{})
			Expect(binding["mode"]).To(Equal("enforcing"))
			Expect(binding["version_number"]).To(Equal(float64(1)))
			Expect(binding["adjustment_number"]).To(Equal(float64(1)))
			Expect(binding["approval_ruleset"]).ToNot(BeNil())

			ruleset := binding["approval_ruleset"].(map[string]interface{})
			Expect(ruleset["id"]).To(Equal("ruleset1"))
			Expect(ruleset["latest_approved_version"]).ToNot(BeNil())
			version := ruleset["latest_approved_version"].(map[string]interface{})
			Expect(version["version_number"]).To(Equal(float64(1)))
		})
	})
})

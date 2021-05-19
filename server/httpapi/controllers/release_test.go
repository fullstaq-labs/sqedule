package controllers

import (
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/gin-gonic/gin"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("approval-ruleset API", func() {
	var ctx HTTPTestContext
	var err error

	type MultipleAppsAndReleasesTestContext struct {
		app1, app2                   dbmodels.Application
		release1, release2, release3 dbmodels.Release
		binding                      dbmodels.ReleaseApprovalRulesetBinding
	}

	SetupMultipleAppsAndReleasesTestContext := func() MultipleAppsAndReleasesTestContext {
		var mctx MultipleAppsAndReleasesTestContext
		var err error

		err = ctx.Db.Transaction(func(tx *gorm.DB) error {
			mctx.app1, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			mctx.app2, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org,
				func(app *dbmodels.Application) {
					app.ID = "app2"
				},
				func(adjustment *dbmodels.ApplicationAdjustment) {
					adjustment.DisplayName = "App 2"
				})
			Expect(err).ToNot(HaveOccurred())

			mctx.release1, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, mctx.app1,
				func(release *dbmodels.Release) {
					release.CreatedAt = time.Now().Add(-1 * time.Second)
				})
			Expect(err).ToNot(HaveOccurred())

			mctx.release2, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, mctx.app2,
				func(release *dbmodels.Release) {
					release.CreatedAt = time.Now().Add(-2 * time.Second)
				})
			Expect(err).ToNot(HaveOccurred())

			mctx.release3, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, mctx.app2,
				func(release *dbmodels.Release) {
					release.CreatedAt = time.Now().Add(-3 * time.Second)
				})
			Expect(err).ToNot(HaveOccurred())

			ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
			Expect(err).ToNot(HaveOccurred())

			mctx.binding, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, mctx.release2,
				ruleset, *ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
			Expect(err).ToNot(HaveOccurred())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())

		return mctx
	}

	BeforeEach(func() {
		ctx, err = SetupHTTPTestContext()
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("POST /applications/:app_id/releases", func() {
		var app dbmodels.Application
		var body gin.H

		Setup := func(autoProcessReleaseInBackground bool) {
			ctx.HttpCtx.AutoProcessReleaseInBackground = autoProcessReleaseInBackground

			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = dbmodels.CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(ctx.Db, ctx.Org, app)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("POST", fmt.Sprintf("/v1/applications/%s/releases", app.ID), gin.H{})
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(201))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		}

		It("creates a release", func() {
			Setup(false)

			var release dbmodels.Release
			tx := ctx.Db.Take(&release)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(release.ApplicationID).To(Equal(app.ID))
		})

		Specify("the created release is not yet processed", func() {
			Setup(false)

			Expect(body["state"]).To(Equal("in_progress"))
			Expect(body["finalized_at"]).To(BeNil())

			var release dbmodels.Release
			tx := ctx.Db.Take(&release)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
			Expect(release.State).To(Equal(releasestate.InProgress))
			Expect(release.FinalizedAt.Valid).To(BeFalse())
		})

		It("creates approval ruleset bindings", func() {
			Setup(false)

			var release dbmodels.Release
			tx := ctx.Db.Take(&release)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

			bindingsJSON := body["approval_ruleset_bindings"].([]interface{})
			Expect(bindingsJSON).To(HaveLen(2))

			bindings, err := dbmodels.FindAllReleaseApprovalRulesetBindings(ctx.Db, ctx.Org.ID, app.ID, release.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(bindings).To(HaveLen(2))
		})

		It("outputs no application", func() {
			Setup(false)

			Expect(body).ToNot(HaveKey("application"))
		})

		It("creates a ReleaseCreatedEvent and CreationAuditRecord", func() {
			Setup(false)

			var creationEvent dbmodels.ReleaseCreatedEvent
			tx := ctx.Db.Take(&creationEvent)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

			var creationRecord dbmodels.CreationAuditRecord
			tx = ctx.Db.Take(&creationRecord)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

			Expect(creationRecord.OrganizationMemberIP.Valid).To(BeFalse())
			Expect(creationRecord.ServiceAccountName.String).To(Equal(ctx.ServiceAccount.Name))
			Expect(creationRecord.ReleaseCreatedEventID).ToNot(BeNil())
			Expect(*creationRecord.ReleaseCreatedEventID).To(Equal(creationEvent.ID))
		})

		It("creates a ReleaseBackgroundJob", func() {
			Setup(false)

			var job dbmodels.ReleaseBackgroundJob
			tx := ctx.Db.Take(&job)
			Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
		})

		It("processes the release eventually", func() {
			Setup(true)
			Eventually(func() releasestate.State {
				var release dbmodels.Release
				tx := ctx.Db.Take(&release)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())
				return release.State
			}).ShouldNot(Equal(releasestate.InProgress))
		})
	})

	Describe("GET /releases", func() {
		var mctx MultipleAppsAndReleasesTestContext
		var body gin.H

		BeforeEach(func() {
			mctx = SetupMultipleAppsAndReleasesTestContext()

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/releases"), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs all releases", func() {
			Expect(body["items"]).To(HaveLen(3))
			items := body["items"].([]interface{})

			Expect(items[0]).To(HaveKey("id"))
			Expect(items[1]).To(HaveKey("id"))
			Expect(items[2]).To(HaveKey("id"))

			Expect(items[0]).To(HaveKey("state"))
			Expect(items[1]).To(HaveKey("state"))
			Expect(items[2]).To(HaveKey("state"))
		})

		It("outputs no approval ruleset bindings", func() {
			Expect(body["items"]).To(HaveLen(3))
			items := body["items"].([]interface{})

			Expect(items[0]).ToNot(HaveKey("approval_ruleset_bindings"))
			Expect(items[1]).ToNot(HaveKey("approval_ruleset_bindings"))
			Expect(items[2]).ToNot(HaveKey("approval_ruleset_bindings"))
		})

		It("outputs applications", func() {
			var app map[string]interface{}

			Expect(body["items"]).To(HaveLen(3))
			items := body["items"].([]interface{})

			item1 := items[0].(map[string]interface{})
			Expect(item1["application"]).ToNot(BeNil())
			app = item1["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", mctx.app1.ID))

			item2 := items[1].(map[string]interface{})
			Expect(item2["application"]).ToNot(BeNil())
			app = item2["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", mctx.app2.ID))

			item3 := items[2].(map[string]interface{})
			Expect(item3["application"]).ToNot(BeNil())
			app = item3["application"].(map[string]interface{})
			Expect(app).To(HaveKeyWithValue("id", mctx.app2.ID))
		})
	})

	Describe("GET /applications/:app_id/releases", func() {
		var mctx MultipleAppsAndReleasesTestContext
		var body gin.H

		BeforeEach(func() {
			mctx = SetupMultipleAppsAndReleasesTestContext()

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s/releases", mctx.app2.ID), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs all releases for that app", func() {
			Expect(body["items"]).To(HaveLen(2))
			items := body["items"].([]interface{})

			Expect(items[0]).To(HaveKeyWithValue("id", float64(mctx.release2.ID)))
			Expect(items[1]).To(HaveKeyWithValue("id", float64(mctx.release3.ID)))
		})

		It("outputs no approval ruleset bindings", func() {
			Expect(body["items"]).To(HaveLen(2))
			items := body["items"].([]interface{})

			item1 := items[0].(map[string]interface{})
			Expect(item1).ToNot(HaveKey("approval_ruleset_bindings"))

			item2 := items[1].(map[string]interface{})
			Expect(item2).ToNot(HaveKey("approval_ruleset_bindings"))
		})

		It("outputs no applications", func() {
			Expect(body["items"]).To(HaveLen(2))
			items := body["items"].([]interface{})

			Expect(items[0]).ToNot(HaveKey("application"))
			Expect(items[1]).ToNot(HaveKey("application"))
		})
	})

	Describe("GET /releases/:id", func() {
		var app dbmodels.Application
		var release dbmodels.Release
		var body gin.H

		BeforeEach(func() {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(ctx.Db, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				release, err = dbmodels.CreateMockReleaseWithInProgressState(ctx.Db, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockRulesetWith1Version(ctx.Db, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode1Version(ctx.Db, ctx.Org, release,
					ruleset, *ruleset.LatestVersion, *ruleset.LatestAdjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s/releases/%d", app.ID, release.ID), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.HttpRecorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs a release", func() {
			Expect(body["state"]).To(Equal("in_progress"))
			Expect(body["finalized_at"]).To(BeNil())
		})

		It("outputs no application", func() {
			Expect(body).ToNot(HaveKey("application"))
		})

		It("outputs approval ruleset bindings", func() {
			Expect(body["approval_ruleset_bindings"]).ToNot(BeEmpty())

			bindings := body["approval_ruleset_bindings"].([]interface{})
			Expect(bindings).To(HaveLen(1))

			binding := bindings[0].(map[string]interface{})
			Expect(binding["mode"]).To(Equal("enforcing"))
			Expect(binding["approval_ruleset"]).ToNot(BeNil())
		})
	})
})

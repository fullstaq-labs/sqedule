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

var _ = Describe("release API", func() {
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

		ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
			mctx.app1, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			mctx.app2, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org,
				func(app *dbmodels.Application) {
					app.ID = "app2"
				},
				func(adjustment *dbmodels.ApplicationAdjustment) {
					adjustment.DisplayName = "App 2"
				})
			Expect(err).ToNot(HaveOccurred())

			mctx.release1, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, mctx.app1,
				func(release *dbmodels.Release) {
					release.CreatedAt = time.Now().Add(-1 * time.Second)
				})
			Expect(err).ToNot(HaveOccurred())

			mctx.release2, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, mctx.app2,
				func(release *dbmodels.Release) {
					release.CreatedAt = time.Now().Add(-2 * time.Second)
				})
			Expect(err).ToNot(HaveOccurred())

			mctx.release3, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, mctx.app2,
				func(release *dbmodels.Release) {
					release.CreatedAt = time.Now().Add(-3 * time.Second)
				})
			Expect(err).ToNot(HaveOccurred())

			ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
			Expect(err).ToNot(HaveOccurred())

			mctx.binding, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, mctx.release2,
				ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
			Expect(err).ToNot(HaveOccurred())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())

		return mctx
	}

	Describe("POST /applications/:app_id/releases", func() {
		var app dbmodels.Application
		var body gin.H

		Setup := func(autoProcessReleaseInBackground bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = dbmodels.CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(tx, ctx.Org, app)
				Expect(err).ToNot(HaveOccurred())

				ctx.ControllerCtx.AutoProcessReleaseInBackground = autoProcessReleaseInBackground

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("POST", fmt.Sprintf("/v1/applications/%s/releases", app.ID), gin.H{})
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.Recorder.Code).To(Equal(201))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		}

		AfterEach(func() {
			ctx.ControllerCtx.WaitGroup.Wait()
		})

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

			req, err := ctx.NewRequestWithAuth("GET", "/v1/releases", nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.Recorder.Code).To(Equal(200))
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

			Expect(ctx.Recorder.Code).To(Equal(200))
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

	Describe("GET /applications/:app_id/releases/:id", func() {
		var app dbmodels.Application
		var release dbmodels.Release
		var body gin.H

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				release, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseRulesetBindingWithEnforcingMode(tx, ctx.Org, release,
					ruleset, *ruleset.Version, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s/releases/%d", app.ID, release.ID), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.Recorder.Code).To(Equal(200))
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
			Expect(body["approval_ruleset_bindings"]).ToNot(BeNil())

			bindings := body["approval_ruleset_bindings"].([]interface{})
			Expect(bindings).To(HaveLen(1))

			binding := bindings[0].(map[string]interface{})
			Expect(binding["mode"]).To(Equal("enforcing"))
			Expect(binding["approval_ruleset"]).To(HaveKeyWithValue("latest_approved_version", Not(BeNil())))

			approvalRuleset := binding["approval_ruleset"].(map[string]interface{})
			approvalRulesetVersion := approvalRuleset["latest_approved_version"]
			Expect(approvalRulesetVersion).To(HaveKeyWithValue("display_name", "Ruleset"))
		})
	})

	Describe("GET /applications/:app_id/releases/:id/events", func() {
		var app dbmodels.Application
		var release dbmodels.Release
		var event1, event2 dbmodels.ReleaseRuleProcessedEvent
		var rule1, rule2 dbmodels.ScheduleApprovalRule
		var outcome1, outcome2 dbmodels.ScheduleApprovalRuleOutcome
		var body gin.H

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				release, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.Org, app, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())
				rule1, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org, ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())
				rule2, err = dbmodels.CreateMockScheduleApprovalRuleWholeDay(tx, ctx.Org, ruleset.Version.ID, *ruleset.Version.Adjustment, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseCreatedEvent(tx, release, nil)
				Expect(err).ToNot(HaveOccurred())

				event1, err = dbmodels.CreateMockReleaseRuleProcessedEvent(tx, release, releasestate.InProgress,
					func(event *dbmodels.ReleaseRuleProcessedEvent) {
						event.CreatedAt = time.Now().Add(-2 * time.Minute)
					})
				Expect(err).ToNot(HaveOccurred())
				outcome1, err = dbmodels.CreateMockScheduleApprovalRuleOutcome(tx, event1, rule1, true, nil)
				Expect(err).ToNot(HaveOccurred())

				event2, err = dbmodels.CreateMockReleaseRuleProcessedEvent(tx, release, releasestate.Cancelled,
					func(event *dbmodels.ReleaseRuleProcessedEvent) {
						event.CreatedAt = time.Now().Add(-1 * time.Minute)
					})
				Expect(err).ToNot(HaveOccurred())
				outcome2, err = dbmodels.CreateMockScheduleApprovalRuleOutcome(tx, event2, rule2, false, nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockReleaseCancelledEvent(tx, release, nil)
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())

			req, err := ctx.NewRequestWithAuth("GET", fmt.Sprintf("/v1/applications/%s/releases/%d/events", app.ID, release.ID), nil)
			Expect(err).ToNot(HaveOccurred())
			ctx.ServeHTTP(req)

			Expect(ctx.Recorder.Code).To(Equal(200))
			body, err = ctx.BodyJSON()
			Expect(err).ToNot(HaveOccurred())
		})

		It("outputs events", func() {
			Expect(body).To(HaveKeyWithValue("items", HaveLen(4)))
			items := body["items"].([]interface{})

			createdEvent := items[0].(map[string]interface{})
			Expect(createdEvent).To(HaveKeyWithValue("type", "created"))
			Expect(createdEvent).To(HaveKey("id"))

			event1JSON := items[1].(map[string]interface{})
			Expect(event1JSON).To(HaveKeyWithValue("type", "rule_processed"))
			Expect(event1JSON).To(HaveKeyWithValue("id", BeNumerically("==", event1.ID)))
			Expect(event1JSON).To(HaveKeyWithValue("result_state", "in_progress"))
			Expect(event1JSON).To(HaveKeyWithValue("ignored_error", BeFalse()))
			Expect(event1JSON).To(HaveKeyWithValue("approval_rule_outcome", Not(BeNil())))
			outcome1JSON := event1JSON["approval_rule_outcome"].(map[string]interface{})
			Expect(outcome1JSON).To(HaveKeyWithValue("type", "schedule"))
			Expect(outcome1JSON).To(HaveKeyWithValue("id", BeNumerically("==", outcome1.ID)))
			Expect(outcome1JSON).To(HaveKeyWithValue("success", BeTrue()))
			Expect(outcome1JSON).To(HaveKeyWithValue("rule", Not(BeNil())))
			rule1JSON := outcome1JSON["rule"].(map[string]interface{})
			Expect(rule1JSON).To(HaveKeyWithValue("id", BeNumerically("==", rule1.ID)))
			Expect(rule1JSON).To(HaveKeyWithValue("begin_time", rule1.BeginTime.String))
			Expect(rule1JSON).To(HaveKeyWithValue("end_time", rule1.EndTime.String))

			event2JSON := items[2].(map[string]interface{})
			Expect(event2JSON).To(HaveKeyWithValue("type", "rule_processed"))
			Expect(event2JSON).To(HaveKeyWithValue("id", BeNumerically("==", event2.ID)))
			Expect(event2JSON).To(HaveKeyWithValue("result_state", "cancelled"))
			Expect(event2JSON).To(HaveKeyWithValue("ignored_error", BeFalse()))
			Expect(event2JSON).To(HaveKeyWithValue("approval_rule_outcome", Not(BeNil())))
			outcome2JSON := event2JSON["approval_rule_outcome"].(map[string]interface{})
			Expect(outcome2JSON).To(HaveKeyWithValue("type", "schedule"))
			Expect(outcome2JSON).To(HaveKeyWithValue("id", BeNumerically("==", outcome2.ID)))
			Expect(outcome2JSON).To(HaveKeyWithValue("success", BeFalse()))
			Expect(outcome2JSON).To(HaveKeyWithValue("rule", Not(BeNil())))
			rule2JSON := outcome2JSON["rule"].(map[string]interface{})
			Expect(rule2JSON).To(HaveKeyWithValue("id", BeNumerically("==", rule2.ID)))
			Expect(rule2JSON).To(HaveKeyWithValue("begin_time", rule2.BeginTime.String))
			Expect(rule2JSON).To(HaveKeyWithValue("end_time", rule2.EndTime.String))

			cancelledEvent := items[3].(map[string]interface{})
			Expect(cancelledEvent).To(HaveKeyWithValue("type", "cancelled"))
			Expect(cancelledEvent).To(HaveKey("id"))
		})
	})
})

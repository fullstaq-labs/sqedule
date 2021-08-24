package controllers

import (
	"fmt"
	"reflect"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/proposalstate"
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
			Path:        "/v1/applications",
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

	Describe("GET /applications/:id/versions", func() {
		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplication(tx, ctx.Org, nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					// Create a binding with 3 versions
					version1, err := dbmodels.CreateMockApplicationVersion(tx, app, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version1, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					// We deliberately create version 3 out of order so that we test
					// whether the versions are outputted in order.

					version3, err := dbmodels.CreateMockApplicationVersion(tx, app, lib.NewUint32Ptr(3), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version3, 1, nil)
					Expect(err).ToNot(HaveOccurred())

					version2, err := dbmodels.CreateMockApplicationVersion(tx, app, lib.NewUint32Ptr(2), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version2, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err := dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version, 1,
						func(adjustment *dbmodels.ApplicationAdjustment) {
							adjustment.ProposalState = proposalstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableListVersionsTest(ReviewableListVersionsTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/applications/app1/versions",
			Setup:       Setup,
		})
	})

	Describe("GET /applications/:id/versions/:version_number", func() {
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
			Path:        "/v1/applications/app1/versions/1",
			Setup:       Setup,

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "app1"))
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

	Describe("GET /applications/:id/proposals", func() {
		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplication(tx, ctx.Org, nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					version, err := dbmodels.CreateMockApplicationVersion(tx, app, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err := dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version, 1,
						func(adjustment *dbmodels.ApplicationAdjustment) {
							adjustment.ProposalState = proposalstate.Draft
						})
					Expect(err).ToNot(HaveOccurred())
				}

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableListProposalsTest(ReviewableListProposalsTestOptions{
			HTTPTestCtx: &ctx,
			Path:        "/v1/applications/app1/proposals",
			Setup:       Setup,
		})
	})

	Describe("GET /applications/:id/proposals/:version_id", func() {
		var version dbmodels.ApplicationVersion

		Setup := func(versionIsApproved bool) {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplication(tx, ctx.Org, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				if versionIsApproved {
					version, err = dbmodels.CreateMockApplicationVersion(tx, app, lib.NewUint32Ptr(1), nil)
					Expect(err).ToNot(HaveOccurred())
					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version, 1, nil)
					Expect(err).ToNot(HaveOccurred())
				} else {
					version, err = dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
					Expect(err).ToNot(HaveOccurred())

					_, err = dbmodels.CreateMockApplicationAdjustment(tx, version, 1,
						func(adjustment *dbmodels.ApplicationAdjustment) {
							adjustment.ProposalState = proposalstate.Draft
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
				return fmt.Sprintf("/v1/applications/app1/proposals/%d", version.ID)
			},
			Setup: Setup,

			ResourceTypeNameInResponse: "application proposal",

			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "app1"))
			},
		})

		It("outputs approval ruleset bindings", func() {
			Setup(false)
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

	Describe("PATCH /applications/:id/proposals/:version_id", func() {
		var app dbmodels.Application
		var proposal1, proposal2, version dbmodels.ApplicationVersion

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Setup := func(hasApprovedVersion bool, proposal1State proposalstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				if hasApprovedVersion {
					app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
					version = *app.Version
				} else {
					app, err = dbmodels.CreateMockApplication(tx, ctx.Org, nil)
				}
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1, err = dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1Adjustment, err := dbmodels.CreateMockApplicationAdjustment(tx, proposal1, 1,
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.ProposalState = proposal1State
					})
				Expect(err).ToNot(HaveOccurred())
				proposal1.Adjustment = &proposal1Adjustment

				proposal2, err = dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal2Adjustment, err := dbmodels.CreateMockApplicationAdjustment(tx, proposal2, 1,
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.ProposalState = proposalstate.Reviewing
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
				return fmt.Sprintf("/v1/applications/app1/proposals/%d", proposal1.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/applications/app1/proposals/%d", version.ID)
			},
			Setup:                      Setup,
			Input:                      gin.H{"display_name": "Changed Name"},
			AutoApproveInput:           gin.H{"display_name": "Changed Name"},
			AdjustmentType:             reflect.TypeOf(dbmodels.ApplicationAdjustment{}),
			ResourceTypeNameInResponse: "application proposal",
			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "app1"))
			},
			GetResourceVersionAndLatestAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				version, err := dbmodels.FindApplicationVersionByID(ctx.Db, ctx.Org.ID, app.ID, proposal1.ID)
				Expect(err).ToNot(HaveOccurred())

				dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApplicationVersion{&version})
				Expect(err).ToNot(HaveOccurred())

				return &version, version.Adjustment
			},
			VersionedFieldJSONFieldName: "display_name",
			VersionedFieldUpdatedValue:  "Changed Name",
			GetSecondProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApplicationVersion
				tx := ctx.Db.Where("id = ?", proposal2.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, ctx.Org.ID,
					[]*dbmodels.ApplicationVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
		})

		It("outputs approval ruleset bindings", func() {
			Setup(true, proposalstate.Draft)
			body := includedTestCtx.MakeRequest(false, false, "", 200)

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

	Describe("PUT /applications/:application_id/:ruleset_id/proposals/:version_id/state", func() {
		var proposal1, proposal2, version dbmodels.ApplicationVersion

		BeforeEach(func() {
			ctx, err = SetupHTTPTestContext(nil)
			Expect(err).ToNot(HaveOccurred())
		})

		Setup := func(proposalState proposalstate.State) {
			err = ctx.Db.Transaction(func(tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				ruleset, err := dbmodels.CreateMockApprovalRulesetWith1Version(tx, ctx.Org, "ruleset1", nil)
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.Org, app, ruleset, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1, err = dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal1Adjustment, err := dbmodels.CreateMockApplicationAdjustment(tx, proposal1, 1,
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.ProposalState = proposalState
					})
				Expect(err).ToNot(HaveOccurred())
				proposal1.Adjustment = &proposal1Adjustment

				proposal2, err = dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				proposal2Adjustment, err := dbmodels.CreateMockApplicationAdjustment(tx, proposal2, 1,
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.ProposalState = proposalState
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
				return fmt.Sprintf("/v1/applications/app1/proposals/%d/state", proposal1.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/applications/app1/proposals/%d/state", version.ID)
			},
			Setup:                      Setup,
			ResourceTypeNameInResponse: "application proposal",
			GetFirstProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApplicationVersion
				tx := ctx.Db.Where("id = ?", proposal1.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApplicationVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
			GetSecondProposalAndAdjustment: func() (dbmodels.IReviewableVersion, dbmodels.IReviewableAdjustment) {
				var proposal dbmodels.ApplicationVersion
				tx := ctx.Db.Where("id = ?", proposal2.ID).Take(&proposal)
				Expect(dbutils.CreateFindOperationError(tx)).ToNot(HaveOccurred())

				err := dbmodels.LoadApplicationVersionsLatestAdjustments(ctx.Db, ctx.Org.ID, []*dbmodels.ApplicationVersion{&proposal})
				Expect(err).ToNot(HaveOccurred())

				return &proposal, proposal.Adjustment
			},
			AssertNonVersionedJSONFieldsExist: func(resource map[string]interface{}) {
				Expect(resource).To(HaveKeyWithValue("id", "app1"))
			},
			VersionedFieldJSONFieldName: "display_name",
			VersionedFieldInitialValue:  "App 1",
		})

		It("outputs approval ruleset bindings", func() {
			Setup(proposalstate.Reviewing)
			body := includedTestCtx.MakeRequest(false, "approved", 200)

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

	Describe("DELETE /applications/:id/proposals/:version_id", func() {
		var version, proposal dbmodels.ApplicationVersion

		Setup := func() {
			ctx, err = SetupHTTPTestContext(func(ctx *HTTPTestContext, tx *gorm.DB) error {
				app, err := dbmodels.CreateMockApplicationWith1Version(tx, ctx.Org, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				// Create a proposal with 2 adjustments

				proposal, err = dbmodels.CreateMockApplicationVersion(tx, app, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				// Adjustment 1

				adjustment1, err := dbmodels.CreateMockApplicationAdjustment(tx, proposal, 1,
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.ProposalState = proposalstate.Draft
					})
				Expect(err).ToNot(HaveOccurred())

				_, err = dbmodels.CreateMockCreationAuditRecord(tx, ctx.Org,
					func(record *dbmodels.CreationAuditRecord) {
						record.ApplicationVersionID = &proposal.ID
						record.ApplicationAdjustmentNumber = &adjustment1.AdjustmentNumber
					})
				Expect(err).ToNot(HaveOccurred())

				// Adjustment 2

				adjustment2, err := dbmodels.CreateMockApplicationAdjustment(tx, proposal, 2,
					func(adjustment *dbmodels.ApplicationAdjustment) {
						adjustment.ProposalState = proposalstate.Draft
					})
				Expect(err).ToNot(HaveOccurred())
				proposal.Adjustment = &adjustment2

				_, err = dbmodels.CreateMockCreationAuditRecord(tx, ctx.Org,
					func(record *dbmodels.CreationAuditRecord) {
						record.ApplicationVersionID = &proposal.ID
						record.ApplicationAdjustmentNumber = &adjustment2.AdjustmentNumber
					})
				Expect(err).ToNot(HaveOccurred())

				return nil
			})
			Expect(err).ToNot(HaveOccurred())
		}

		IncludeReviewableDeleteProposalTest(ReviewableDeleteProposalTestOptions{
			HTTPTestCtx: &ctx,
			GetProposalPath: func() string {
				return fmt.Sprintf("/v1/applications/app1/proposals/%d", proposal.ID)
			},
			GetApprovedVersionPath: func() string {
				return fmt.Sprintf("/v1/applications/app1/proposals/%d", version.ID)
			},
			Setup:                      Setup,
			ResourceTypeNameInResponse: "application proposal",
			CountProposals: func() uint {
				var count int64
				err = ctx.Db.Model(dbmodels.ApplicationVersion{}).Where("version_number IS NULL").Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				return uint(count)
			},
			CountProposalAdjustments: func() uint {
				var count int64
				err = ctx.Db.Model(dbmodels.ApplicationAdjustment{}).Where("application_version_id = ?", proposal.ID).Count(&count).Error
				Expect(err).ToNot(HaveOccurred())
				return uint(count)
			},
		})
	})
})

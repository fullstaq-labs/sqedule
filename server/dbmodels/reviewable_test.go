package dbmodels

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/lib"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadReviewablesLatestVersionsAndAdjustments", func() {
	var db *gorm.DB
	var err error
	var org Organization
	var app Application
	var ruleset1 ApprovalRuleset
	var ruleset2 ApprovalRuleset

	BeforeEach(func() {
		db, err = dbutils.SetupTestDatabase()
		Expect(err).ToNot(HaveOccurred())

		err = db.Transaction(func(tx *gorm.DB) error {
			org, err = CreateMockOrganization(tx, nil)
			Expect(err).ToNot(HaveOccurred())
			app, err = CreateMockApplicationWith1Version(tx, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			ruleset1, err = CreateMockApprovalRulesetWith1Version(tx, org, "ruleset1", nil)
			Expect(err).ToNot(HaveOccurred())
			ruleset2, err = CreateMockApprovalRulesetWith1Version(tx, org, "ruleset2", nil)
			Expect(err).ToNot(HaveOccurred())
			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("works", func() {
		var versionNumber2 uint32 = 2
		err = db.Transaction(func(tx *gorm.DB) error {
			// Create binding 1
			binding1, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, org, app,
				ruleset1, nil)
			Expect(err).ToNot(HaveOccurred())

			// Binding 1: create version 2.1 and 2.2
			binding1Version2, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding1, &versionNumber2)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding1Version2, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding1Version2, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
				adjustment.AdjustmentNumber = 2
			})
			Expect(err).ToNot(HaveOccurred())

			// Binding 1: Create next version (no version number) and its adjustment
			binding1VersionNext, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding1, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding1VersionNext, nil)
			Expect(err).ToNot(HaveOccurred())

			// Create binding 2
			binding2, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, org, app,
				ruleset2, nil)
			Expect(err).ToNot(HaveOccurred())

			// Binding 2: create version 2.1, 2.2 and 2.3
			binding2Version2, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding2, &versionNumber2)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding2Version2, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding2Version2, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
				adjustment.AdjustmentNumber = 2
			})
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding2Version2, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
				adjustment.AdjustmentNumber = 3
			})
			Expect(err).ToNot(HaveOccurred())

			// Binding 2: Create next version (no version number) and its adjustment
			binding2VersionNext, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding2, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, org, binding2VersionNext, nil)
			Expect(err).ToNot(HaveOccurred())

			err = LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding1, &binding2})
			Expect(err).ToNot(HaveOccurred())

			// Run test: binding1's latest version should be 2.2

			Expect(binding1.Version).ToNot(BeNil())
			Expect(binding1.Version.VersionNumber).ToNot(BeNil())
			Expect(binding1.Version.Adjustment).ToNot(BeNil())
			Expect(*binding1.Version.VersionNumber).To(BeNumerically("==", 2))
			Expect(binding1.Version.Adjustment.AdjustmentNumber).To(BeNumerically("==", 2))

			// Run test: binding2's latest version should be 2.3
			Expect(binding2.Version).ToNot(BeNil())
			Expect(binding2.Version.VersionNumber).ToNot(BeNil())
			Expect(binding2.Version.Adjustment).ToNot(BeNil())
			Expect(*binding2.Version.VersionNumber).To(BeNumerically("==", 2))
			Expect(binding2.Version.Adjustment.AdjustmentNumber).To(BeNumerically("==", 3))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("loads nothing when there are no versions", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			binding := ApplicationApprovalRulesetBinding{
				BaseModel: BaseModel{
					OrganizationID: org.ID,
				},
				ApplicationApprovalRulesetBindingPrimaryKey: ApplicationApprovalRulesetBindingPrimaryKey{
					ApplicationID:     app.ID,
					ApprovalRulesetID: ruleset1.ID,
				},
			}
			savetx := tx.Create(&binding)
			Expect(savetx.Error).ToNot(HaveOccurred())

			// Run test: latest version should not exist
			err = LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding})
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.Version).To(BeNil())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("loads nothing when there are only proposed versions", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			binding := ApplicationApprovalRulesetBinding{
				BaseModel: BaseModel{
					OrganizationID: org.ID,
				},
				ApplicationApprovalRulesetBindingPrimaryKey: ApplicationApprovalRulesetBindingPrimaryKey{
					ApplicationID:     app.ID,
					ApprovalRulesetID: ruleset1.ID,
				},
			}
			savetx := tx.Create(&binding)
			Expect(savetx.Error).ToNot(HaveOccurred())

			_, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding, nil)
			Expect(err).ToNot(HaveOccurred())

			// Run test: latest version should not exist
			err = LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding})
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.Version).To(BeNil())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("loads no adjustments when there are no adjustments", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, org, app,
				ruleset1, nil)
			Expect(err).ToNot(HaveOccurred())
			binding.Version = nil

			// Create version 2 with no adjustments
			var versionNumber2 uint32 = 2
			_, err = CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding, &versionNumber2)
			Expect(err).ToNot(HaveOccurred())

			// Run test: latest version should be 2, adjustment version nil
			err = LoadApplicationApprovalRulesetBindingsLatestVersionsAndAdjustments(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding})
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.Version).ToNot(BeNil())
			Expect(binding.Version.VersionNumber).ToNot(BeNil())
			Expect(binding.Version.Adjustment).To(BeNil())
			Expect(*binding.Version.VersionNumber).To(BeNumerically("==", 2))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("LoadReviewablesLatestVersions", func() {
	var db *gorm.DB
	var err error
	var org Organization

	BeforeEach(func() {
		db, err = dbutils.SetupTestDatabase()
		Expect(err).ToNot(HaveOccurred())

		org, err = CreateMockOrganization(db, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("works", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			// Create an app with 2 versions

			app1, err := CreateMockApplicationWith1Version(tx, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationVersion(tx, app1, lib.NewUint32Ptr(2), nil)
			Expect(err).ToNot(HaveOccurred())

			// Create an app with 3 versions

			app2, err := CreateMockApplicationWith1Version(tx, org, func(app *Application) {
				app.ID = "app2"
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			// We deliberately create version 3 out of order so that we test
			// whether LoadReviewablesLatestVersions() returns the highest
			// version number.
			_, err = CreateMockApplicationVersion(tx, app2, lib.NewUint32Ptr(3), nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationVersion(tx, app2, lib.NewUint32Ptr(2), nil)
			Expect(err).ToNot(HaveOccurred())

			// Run test

			app1.Version = nil
			app2.Version = nil
			err = LoadApplicationsLatestVersions(tx, org.ID, []*Application{&app1, &app2})
			Expect(err).ToNot(HaveOccurred())
			Expect(*app1.Version.VersionNumber).To(BeNumerically("==", 2))
			Expect(*app2.Version.VersionNumber).To(BeNumerically("==", 3))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("works on resources with composite primary keys", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			app1, err := CreateMockApplication(tx, org, nil)
			Expect(err).ToNot(HaveOccurred())
			app2, err := CreateMockApplication(tx, org, func(app *Application) {
				app.ID = "app2"
			})
			Expect(err).ToNot(HaveOccurred())

			ruleset, err := CreateMockApprovalRulesetWith1Version(tx, org, "ruleset", nil)
			Expect(err).ToNot(HaveOccurred())

			// Create two application ruleset bindings:
			// binding1 has 2 versions
			// binding2 has 3 versions

			binding1, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, org, app1, ruleset, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app1, binding1, lib.NewUint32Ptr(2))
			Expect(err).ToNot(HaveOccurred())

			binding2, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, org, app2, ruleset, nil)
			Expect(err).ToNot(HaveOccurred())
			// We deliberately create version 3 out of order so that we test
			// whether LoadReviewablesLatestVersions() returns the highest
			// version number.
			_, err = CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app2, binding2, lib.NewUint32Ptr(3))
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app2, binding2, lib.NewUint32Ptr(2))
			Expect(err).ToNot(HaveOccurred())

			binding1.Version = nil
			binding2.Version = nil

			// Run test

			err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding1, &binding2})
			Expect(err).ToNot(HaveOccurred())
			Expect(*binding1.Version.VersionNumber).To(BeNumerically("==", 2))
			Expect(*binding2.Version.VersionNumber).To(BeNumerically("==", 3))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("works given multiple Reviewables with the same primary key", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			app, err := CreateMockApplicationWith1Version(tx, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			app.Version = nil
			appCopy := app

			err = LoadApplicationsLatestVersions(tx, org.ID, []*Application{&app, &appCopy})
			Expect(err).ToNot(HaveOccurred())
			Expect(*app.Version.VersionNumber).To(BeNumerically("==", 1))
			Expect(*appCopy.Version.VersionNumber).To(BeNumerically("==", 1))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("does not load proposed versions", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			app, err := CreateMockApplication(tx, org, nil)
			Expect(err).ToNot(HaveOccurred())

			_, err = CreateMockApplicationVersion(tx, app, nil, func(version *ApplicationVersion) {
				version.VersionNumber = nil
				version.ApprovedAt = sql.NullTime{Time: time.Time{}, Valid: false}
			})
			Expect(err).ToNot(HaveOccurred())

			err = LoadApplicationsLatestVersions(tx, org.ID, []*Application{&app})
			Expect(err).ToNot(HaveOccurred())
			Expect(app.Version).To(BeNil())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("loads nothing when there are no versions", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			app, err := CreateMockApplication(tx, org, nil)
			Expect(err).ToNot(HaveOccurred())

			err = LoadApplicationsLatestVersions(tx, org.ID, []*Application{&app})
			Expect(err).ToNot(HaveOccurred())
			Expect(app.Version).To(BeNil())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("LoadReviewableVersionsLatestAdjustments", func() {
	var db *gorm.DB
	var err error
	var org Organization

	BeforeEach(func() {
		db, err = dbutils.SetupTestDatabase()
		Expect(err).ToNot(HaveOccurred())

		org, err = CreateMockOrganization(db, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	It("works", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			// Create an app with 1 version and 2 adjustments

			app1, err := CreateMockApplicationWith1Version(tx, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			app1Adjustment2, err := CreateMockApplicationAdjustment(tx, *app1.Version, 2, nil)
			Expect(err).ToNot(HaveOccurred())
			app1.Version.Adjustment = &app1Adjustment2

			// Create an app with 2 versions:
			// Version 1 has 2 adjustments
			// Version 2 has 3 adjustments

			app2, err := CreateMockApplicationWith1Version(tx, org, func(app *Application) {
				app.ID = "app2"
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			app2Version1Adjustment2, err := CreateMockApplicationAdjustment(tx, *app2.Version, 2, nil)
			Expect(err).ToNot(HaveOccurred())
			app2.Version.Adjustment = &app2Version1Adjustment2

			app2Version2, err := CreateMockApplicationVersion(tx, app2, lib.NewUint32Ptr(2), nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationAdjustment(tx, app2Version2, 1, nil)
			Expect(err).ToNot(HaveOccurred())
			// We deliberately create adjustment 3 out of order so that we test
			// whether LoadReviewableVersionsLatestAdjustments() returns the highest
			// adjustment number.
			app2Version2Adjustment3, err := CreateMockApplicationAdjustment(tx, app2Version2, 3, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationAdjustment(tx, app2Version2, 2, nil)
			Expect(err).ToNot(HaveOccurred())
			app2.Version = &app2Version2
			app2.Version.Adjustment = &app2Version2Adjustment3

			// Run test

			err = LoadApplicationVersionsLatestAdjustments(tx, org.ID, []*ApplicationVersion{app1.Version, app2.Version})
			Expect(err).ToNot(HaveOccurred())
			Expect(app1.Version.Adjustment.AdjustmentNumber).To(BeNumerically("==", 2))
			Expect(app2.Version.Adjustment.AdjustmentNumber).To(BeNumerically("==", 3))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("works given multiple Versions with the same primary key", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			// Create an app with 1 version and 2 adjustments

			app, err := CreateMockApplicationWith1Version(tx, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			_, err = CreateMockApplicationAdjustment(tx, *app.Version, 2, nil)
			Expect(err).ToNot(HaveOccurred())

			versionCopy1 := *app.Version
			versionCopy1.Adjustment = nil
			versionCopy2 := *app.Version
			versionCopy2.Adjustment = nil

			// Run test

			err = LoadApplicationVersionsLatestAdjustments(tx, org.ID, []*ApplicationVersion{&versionCopy1, &versionCopy2})
			Expect(err).ToNot(HaveOccurred())
			Expect(versionCopy1.Adjustment.AdjustmentNumber).To(BeNumerically("==", 2))
			Expect(versionCopy2.Adjustment.AdjustmentNumber).To(BeNumerically("==", 2))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("loads nothing when there are no adjustments", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			app, err := CreateMockApplicationWith1Version(tx, org, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			version2, err := CreateMockApplicationVersion(tx, app, lib.NewUint32Ptr(2), nil)
			Expect(err).ToNot(HaveOccurred())

			err = LoadApplicationVersionsLatestAdjustments(tx, org.ID, []*ApplicationVersion{&version2})
			Expect(err).ToNot(HaveOccurred())
			Expect(version2.Adjustment).To(BeNil())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})
})

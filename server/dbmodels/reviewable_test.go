package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"gorm.io/gorm"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = FDescribe("LoadReviewablesLatestVersions", func() {
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
			ruleset1, err = CreateMockRulesetWith1Version(tx, org, "ruleset1", nil)
			Expect(err).ToNot(HaveOccurred())
			ruleset2, err = CreateMockRulesetWith1Version(tx, org, "ruleset2", nil)
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

			err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding1, &binding2})
			Expect(err).ToNot(HaveOccurred())

			// Run test: binding1's latest version should be 2.2

			Expect(binding1.LatestVersion).ToNot(BeNil())
			Expect(binding1.LatestVersion.VersionNumber).ToNot(BeNil())
			Expect(binding1.LatestAdjustment).ToNot(BeNil())
			Expect(*binding1.LatestVersion.VersionNumber).To(BeNumerically("==", 2))
			Expect(binding1.LatestAdjustment.AdjustmentNumber).To(BeNumerically("==", 2))

			// Run test: binding2's latest version should be 2.3
			Expect(binding2.LatestVersion).ToNot(BeNil())
			Expect(binding2.LatestVersion.VersionNumber).ToNot(BeNil())
			Expect(binding2.LatestAdjustment).ToNot(BeNil())
			Expect(*binding2.LatestVersion.VersionNumber).To(BeNumerically("==", 2))
			Expect(binding2.LatestAdjustment.AdjustmentNumber).To(BeNumerically("==", 3))

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
			err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding})
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.LatestVersion).To(BeNil())
			Expect(binding.LatestAdjustment).To(BeNil())

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
			err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding})
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.LatestVersion).To(BeNil())
			Expect(binding.LatestAdjustment).To(BeNil())

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})

	It("loads no adjustments when there are no adjustments", func() {
		err = db.Transaction(func(tx *gorm.DB) error {
			binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, org, app,
				ruleset1, nil)
			Expect(err).ToNot(HaveOccurred())
			binding.LatestVersion = nil
			binding.LatestAdjustment = nil

			// Create version 2 with no adjustments
			var versionNumber2 uint32 = 2
			_, err = CreateMockApplicationApprovalRulesetBindingVersion(tx, org, app, binding, &versionNumber2)
			Expect(err).ToNot(HaveOccurred())

			// Run test: latest version should be 2, adjustment version nil
			err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, org.ID, []*ApplicationApprovalRulesetBinding{&binding})
			Expect(err).ToNot(HaveOccurred())
			Expect(binding.LatestVersion).ToNot(BeNil())
			Expect(binding.LatestVersion.VersionNumber).ToNot(BeNil())
			Expect(binding.LatestAdjustment).To(BeNil())
			Expect(*binding.LatestVersion.VersionNumber).To(BeNumerically("==", 2))

			return nil
		})
		Expect(err).ToNot(HaveOccurred())
	})
})

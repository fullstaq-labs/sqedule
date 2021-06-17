package dbmodels

import (
	"github.com/fullstaq-labs/sqedule/server/dbutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ApprovalRulesetContents", func() {
	It("supports all ruleset types", func() {
		contents := ApprovalRulesetContents{
			HTTPApiApprovalRules:  []HTTPApiApprovalRule{{}},
			ScheduleApprovalRules: []ScheduleApprovalRule{{}},
			ManualApprovalRules:   []ManualApprovalRule{{}},
		}
		Expect(contents.NumRules()).To(BeNumerically("==", NumApprovalRuleTypes))
	})
})

var _ = Describe("ApprovalRuleset finders", func() {
	Describe("FindApprovalRulesBoundToRelease", func() {
		It("supports all ruleset types", func() {
			db, err := dbutils.SetupTestDatabase()
			Expect(err).ToNot(HaveOccurred())
			Expect(func() {
				_, err = FindApprovalRulesBoundToRelease(db, "org", "app", 0)
			}).ToNot(Panic())
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("LoadApprovalRulesetAdjustmentsApprovalRules", func() {
		It("supports all ruleset types", func() {
			db, err := dbutils.SetupTestDatabase()
			Expect(err).ToNot(HaveOccurred())
			Expect(func() {
				adjustment := ApprovalRulesetAdjustment{}
				err = LoadApprovalRulesetAdjustmentsApprovalRules(db, "org", []*ApprovalRulesetAdjustment{&adjustment})
			}).ToNot(Panic())
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

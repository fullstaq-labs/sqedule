package dbmodels

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/stretchr/testify/assert"
)

func TestApprovalRulesetContentsNumRulesSupportsAllRulesetTypes(t *testing.T) {
	contents := ApprovalRulesetContents{
		HTTPApiApprovalRules:  []HTTPApiApprovalRule{{}},
		ScheduleApprovalRules: []ScheduleApprovalRule{{}},
		ManualApprovalRules:   []ManualApprovalRule{{}},
	}
	assert.Equal(t, uint(NumApprovalRuleTypes), contents.NumRules())
}

func TestFindApprovalRulesBoundToReleaseSupportsAllRulesetTypes(t *testing.T) {
	db, err := dbutils.SetupTestDatabase()
	if !assert.NoError(t, err) {
		return
	}

	assert.NotPanics(t, func() {
		FindApprovalRulesBoundToRelease(db, "org", "app", 0)
	})
}

func TestFindApprovalRulesInRulesetVersionSupportsAllRulesetTypes(t *testing.T) {
	db, err := dbutils.SetupTestDatabase()
	if !assert.NoError(t, err) {
		return
	}

	assert.NotPanics(t, func() {
		FindApprovalRulesInRulesetVersion(db, "org", ApprovalRulesetVersionAndAdjustmentKey{
			VersionID:        1,
			AdjustmentNumber: 1,
		})
	})
}

package approvalrulesengine

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"gorm.io/gorm"
)

func (engine Engine) loadManualApprovalRules(db *gorm.DB, majorVersionIndex map[uint64]*ruleset, versionKeys []dbmodels.ApprovalRulesetVersionKey) (uint, error) {
	// TODO
	//rules, err := dbmodels.FindAllManualApprovalRulesBelongingToVersions(
	//	db, engine.Organization.ID, versionKeys)
	// if err != nil {
	// 	return 0, err
	// }
	rules := make([]dbmodels.ManualApprovalRule, 0)

	for _, rule := range rules {
		ruleset := majorVersionIndex[rule.ApprovalRulesetMajorVersionID]
		ruleset.ManualApprovalRules = append(ruleset.ManualApprovalRules, rule)
	}

	return uint(len(rules)), nil
}

func (engine Engine) fetchManualApprovalRulePreviousOutcomes() (map[uint64]bool, error) {
	// TODO
	// outcomes, err := dbmodels.FindAllManualApprovalRuleOutcomes(engine.Db, engine.Organization.ID, engine.Release.ID)
	// if err != nil {
	// 	return nil, err
	// }
	outcomes := make([]dbmodels.ManualApprovalRuleOutcome, 0)

	return indexManualApprovalRuleOutcomes(outcomes), nil
}

func (engine Engine) processManualApprovalRules(rulesets []ruleset, previousOutcomes map[uint64]bool, nAlreadyProcessed uint, totalRules uint) (releasestate.State, uint, error) {
	var nprocessed uint = 0

	// TODO

	return determineReleaseStateAfterProcessingRules(nAlreadyProcessed, nprocessed, totalRules),
		nprocessed, nil
}

func indexManualApprovalRuleOutcomes(outcomes []dbmodels.ManualApprovalRuleOutcome) map[uint64]bool {
	result := make(map[uint64]bool)
	for _, outcome := range outcomes {
		result[outcome.ManualApprovalRuleID] = outcome.Success
	}
	return result
}

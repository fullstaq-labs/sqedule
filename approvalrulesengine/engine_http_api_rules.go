package approvalrulesengine

import (
	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"gorm.io/gorm"
)

func (engine Engine) loadHTTPApiRules(db *gorm.DB, majorVersionIndex map[uint64]*ruleset, versionKeys []dbmodels.ApprovalRulesetVersionKey) (uint, error) {
	// TODO
	//rules, err := dbmodels.FindAllHTTPApiApprovalRulesBelongingToVersions(
	//	db, engine.Organization.ID, versionKeys)
	// if err != nil {
	// 	return 0, err
	// }
	rules := make([]dbmodels.HTTPApiApprovalRule, 0)

	for _, rule := range rules {
		ruleset := majorVersionIndex[rule.ApprovalRulesetMajorVersionID]
		ruleset.httpAPIRules = append(ruleset.httpAPIRules, rule)
	}

	return uint(len(rules)), nil
}

func (engine Engine) fetchHTTPApiRulePreviousOutcomes() (map[uint64]bool, error) {
	// TODO
	// outcomes, err := dbmodels.FindAllHTTPApiApprovalRuleOutcomes(engine.Db, engine.Organization.ID, engine.Release.ID)
	// if err != nil {
	// 	return nil, err
	// }
	outcomes := make([]dbmodels.HTTPApiApprovalRuleOutcome, 0)

	return indexHTTPApiRuleOutcomes(outcomes), nil
}

func (engine Engine) processHTTPApiRules(rulesets []ruleset, previousOutcomes map[uint64]bool, nAlreadyProcessed uint, totalRules uint) (releasestate.State, uint, error) {
	var nprocessed uint = 0

	// TODO

	return determineReleaseStateAfterProcessingRules(nAlreadyProcessed, nprocessed, totalRules),
		nprocessed, nil
}

func indexHTTPApiRuleOutcomes(outcomes []dbmodels.HTTPApiApprovalRuleOutcome) map[uint64]bool {
	result := make(map[uint64]bool)
	for _, outcome := range outcomes {
		result[outcome.ApprovalRuleOutcome.ID] = outcome.Success
	}
	return result
}

package approvalrulesprocessing

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
)

func (engine Engine) fetchHTTPApiRulePreviousOutcomes() (map[uint64]bool, error) {
	// TODO
	// outcomes, err := dbmodels.FindAllHTTPApiApprovalRuleOutcomes(engine.Db, engine.Organization.ID, engine.Release.ID)
	// if err != nil {
	// 	return nil, err
	// }
	outcomes := make([]dbmodels.HTTPApiApprovalRuleOutcome, 0)

	return indexHTTPApiRuleOutcomes(outcomes), nil
}

func (engine Engine) processHTTPApiRules(rulesetContents dbmodels.ApprovalRulesetContents, previousOutcomes map[uint64]bool, nAlreadyProcessed uint) (releasestate.State, uint, error) {
	var nprocessed uint = 0

	// TODO

	return determineReleaseStateAfterProcessingRules(nAlreadyProcessed, nprocessed, rulesetContents.NumRules()),
		nprocessed, nil
}

func indexHTTPApiRuleOutcomes(outcomes []dbmodels.HTTPApiApprovalRuleOutcome) map[uint64]bool {
	result := make(map[uint64]bool)
	for _, outcome := range outcomes {
		result[outcome.HTTPApiApprovalRuleID] = outcome.Success
	}
	return result
}

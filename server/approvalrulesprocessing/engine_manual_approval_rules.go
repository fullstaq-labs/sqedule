package approvalrulesprocessing

import (
	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
)

func (engine Engine) fetchManualApprovalRulePreviousOutcomes() (map[uint64]bool, error) {
	// TODO
	// outcomes, err := dbmodels.FindAllManualApprovalRuleOutcomes(engine.Db, engine.Organization.ID, engine.Release.ID)
	// if err != nil {
	// 	return nil, err
	// }
	outcomes := make([]dbmodels.ManualApprovalRuleOutcome, 0)

	return indexManualApprovalRuleOutcomes(outcomes), nil
}

func (engine Engine) processManualApprovalRules(rulesetContents dbmodels.ApprovalRulesetContents, previousOutcomes map[uint64]bool, nAlreadyProcessed uint) (releasestate.State, uint, error) {
	var nprocessed uint = 0

	// TODO

	return determineReleaseStateAfterProcessingRules(nAlreadyProcessed, nprocessed, rulesetContents.NumRules()),
		nprocessed, nil
}

func indexManualApprovalRuleOutcomes(outcomes []dbmodels.ManualApprovalRuleOutcome) map[uint64]bool {
	result := make(map[uint64]bool)
	for _, outcome := range outcomes {
		result[outcome.ManualApprovalRuleID] = outcome.Success
	}
	return result
}

package approvalrulesprocessing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"gorm.io/gorm"
)

// Engine processes a Release based on its bound ApprovalRules.
// It accepts a Release which hasn't been fully processed yet. This is
// done by passing the corresponding ReleaseBackgroundJob.
type Engine struct {
	Db                   *gorm.DB
	OrganizationID       string
	ReleaseBackgroundJob dbmodels.ReleaseBackgroundJob
}

var errTemporary = errors.New("temporary error, retry later")

func (engine *Engine) Run() error {
	locktx, err := engine.lock()
	if err != nil {
		return fmt.Errorf("Error acquiring lock: %w", err)
	}
	defer engine.unlock(locktx)

	rulesetContents, err := engine.loadRules()
	if err != nil {
		return fmt.Errorf("Error loading rules: %w", err)
	}

	resultState, err := engine.processRules(rulesetContents)
	if err != nil {
		// Error message already mentions the fact that it's about processing rules.
		return err
	}

	err = engine.finalizeJob(resultState)
	if err != nil {
		return fmt.Errorf("Error finalizing release background job (ID=%d) to state %s: %w",
			engine.ReleaseBackgroundJob.ReleaseID, resultState, err)
	}
	return nil
}

func (engine *Engine) processRules(rulesetContents dbmodels.ApprovalRulesetContents) (releasestate.State, error) {
	var finalResultState releasestate.State = releasestate.InProgress
	var finalError error
	var resultState releasestate.State
	var nprocessed uint = 0
	var n uint

	manualApprovalRulePreviousOutcomes, err := engine.fetchManualApprovalRulePreviousOutcomes()
	if err != nil {
		return releasestate.Rejected, fmt.Errorf("Error loading state: %w", err)
	}
	scheduleRulePreviousOutcomes, err := engine.fetchScheduleRulePreviousOutcomes()
	if err != nil {
		return releasestate.Rejected, fmt.Errorf("Error loading state: %w", err)
	}
	httpAPIRulePreviousOutcomes, err := engine.fetchHTTPApiRulePreviousOutcomes()
	if err != nil {
		return releasestate.Rejected, fmt.Errorf("Error loading state: %w", err)
	}

	err = engine.Db.Transaction(func(tx *gorm.DB) error {
		if !finalResultState.IsFinal() {
			// Process manual approval rules
			resultState, n, err = engine.processManualApprovalRules(rulesetContents, manualApprovalRulePreviousOutcomes, nprocessed)
			if err != nil {
				finalResultState = releasestate.Rejected
				// Error message already mentions the fact that it's about processing rules.
				finalError = err
				return nil
			}
			nprocessed += n
			if resultState.IsFinal() {
				finalResultState = resultState
			}
		}

		if !finalResultState.IsFinal() {
			// Process schedule rules
			resultState, n, err = engine.processScheduleRules(rulesetContents, scheduleRulePreviousOutcomes, nprocessed)
			if err != nil {
				finalResultState = releasestate.Rejected
				// Error message already mentions the fact that it's about processing rules.
				finalError = err
				return nil
			}
			nprocessed += n
			if resultState.IsFinal() {
				finalResultState = resultState
			}
		}

		return nil
	})
	if err != nil {
		return releasestate.Rejected, err
	}
	if finalError != nil {
		return releasestate.Rejected, finalError
	}

	if !finalResultState.IsFinal() {
		// Process HTTP API rules
		resultState, n, err = engine.processHTTPApiRules(rulesetContents, httpAPIRulePreviousOutcomes, nprocessed)
		if err != nil {
			// Error message already mentions the fact that it's about processing rules.
			return releasestate.Rejected, err
		}
		nprocessed += n
		if resultState.IsFinal() {
			finalResultState = resultState
		}
	}

	if !finalResultState.IsFinal() {
		panic("Bug: none of the rule processors returned a final result state")
	}
	return finalResultState, finalError
}

func (engine Engine) lock() (*gorm.DB, error) {
	// Go's database connections are automatically pooled. We reserve a transaction here (to be passed to unlock())
	// in order to ensure that we release the lock from the same database connection.
	tx := engine.Db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("Error starting a transaction: %w", tx.Error)
	}

	tx2 := tx.Exec("SELECT pg_advisory_lock(?) AS result", engine.getPostgresAdvisoryLockID())
	if tx2.Error != nil {
		return nil, tx2.Error
	}

	return tx, nil
}

func (engine Engine) unlock(locktx *gorm.DB) {
	var result struct {
		Result bool
	}
	tx := locktx.Raw("SELECT pg_advisory_unlock(?) AS result", engine.getPostgresAdvisoryLockID()).Scan(&result)
	if tx.Error != nil {
		engine.Db.Logger.Warn(context.Background(), "Error releasing advisory lock %d: %s",
			engine.getPostgresAdvisoryLockID(), tx.Error.Error())
	}

	if !result.Result {
		engine.Db.Logger.Warn(context.Background(), "Error releasing advisory lock %d: database returned false",
			engine.getPostgresAdvisoryLockID())
	}
}

func (engine Engine) getPostgresAdvisoryLockID() uint64 {
	return dbmodels.ReleaseBackgroundJobPostgresLockNamespace + uint64(engine.ReleaseBackgroundJob.LockSubID)
}

func (engine Engine) loadRules() (dbmodels.ApprovalRulesetContents, error) {
	return dbmodels.FindApprovalRulesBoundToRelease(engine.Db, engine.OrganizationID,
		engine.ReleaseBackgroundJob.ApplicationID, engine.ReleaseBackgroundJob.ReleaseID)
}

func (engine *Engine) finalizeJob(resultState releasestate.State) error {
	return engine.Db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		release := &engine.ReleaseBackgroundJob.Release
		release.State = resultState
		release.FinalizedAt = sql.NullTime{Time: now, Valid: true}
		savetx := tx.Model(release).Updates(map[string]interface{}{
			"state":        resultState,
			"finalized_at": now,
		})
		if savetx.Error != nil {
			return savetx.Error
		}

		return tx.Delete(&engine.ReleaseBackgroundJob).Error
	})
}

func (engine Engine) createRuleProcessedEvent(resultState releasestate.State, ignoredError bool) (dbmodels.ReleaseRuleProcessedEvent, error) {
	event := dbmodels.ReleaseRuleProcessedEvent{
		ReleaseEvent: dbmodels.ReleaseEvent{
			BaseModel: dbmodels.BaseModel{
				OrganizationID: engine.OrganizationID,
			},
			ReleaseID:     engine.ReleaseBackgroundJob.Release.ID,
			ApplicationID: engine.ReleaseBackgroundJob.ApplicationID,
		},
		ResultState:  resultState,
		IgnoredError: ignoredError,
	}
	tx := engine.Db.Create(&event)
	if tx.Error != nil {
		return dbmodels.ReleaseRuleProcessedEvent{}, tx.Error
	}

	return event, nil
}

func maybeFormatRuleProcessingError(err error, format string, a ...interface{}) error {
	if errors.Is(err, errTemporary) {
		return err
	}
	return fmt.Errorf(format, a...)
}

func isLastRule(nAlreadyProcessed uint, nprocessed uint, totalRules uint) bool {
	return nAlreadyProcessed+nprocessed == totalRules
}

func determineReleaseStateFromOutcome(ruleProcessedSuccessfully bool, mode approvalrulesetbindingmode.Mode, isLastRule bool) (state releasestate.State, ignoredError bool) {
	if ruleProcessedSuccessfully || mode == approvalrulesetbindingmode.Permissive {
		if isLastRule {
			return releasestate.Approved, !ruleProcessedSuccessfully
		}
		return releasestate.InProgress, !ruleProcessedSuccessfully
	}
	return releasestate.Rejected, false
}

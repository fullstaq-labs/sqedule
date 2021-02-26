package approvalrulesengine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
	"gorm.io/gorm"
)

func (engine Engine) loadScheduleRules(db *gorm.DB, majorVersionIndex map[uint64]*ruleset, versionKeys []dbmodels.ApprovalRuleVersionKey) (uint, error) {
	rules, err := dbmodels.FindAllScheduleApprovalRulesBelongingToVersions(
		db, engine.Organization.ID, versionKeys)
	if err != nil {
		return 0, err
	}

	for _, rule := range rules {
		ruleset := majorVersionIndex[rule.ApprovalRulesetMajorVersionID]
		ruleset.scheduleRules = append(ruleset.scheduleRules, rule)
	}

	return uint(len(rules)), nil
}

func (engine Engine) fetchScheduleRulePreviousOutcomes() (map[uint64]bool, error) {
	outcomes, err := dbmodels.FindAllScheduleApprovalRuleOutcomes(engine.Db, engine.Organization.ID, engine.ReleaseBackgroundJob.DeploymentRequestID)
	if err != nil {
		return nil, err
	}

	return indexScheduleRuleOutcomes(outcomes), nil
}

func (engine Engine) processScheduleRules(rulesets []ruleset, previousOutcomes map[uint64]bool, nAlreadyProcessed uint, totalRules uint) (deploymentrequeststate.State, uint, error) {
	var nprocessed uint = 0

	for _, ruleset := range rulesets {
		for _, rule := range ruleset.scheduleRules {
			success, outcomeAlreadyRecorded, err := engine.processScheduleRule(rule, previousOutcomes)
			if err != nil {
				return deploymentrequeststate.Rejected, nprocessed,
					maybeFormatRuleProcessingError(err, "Error processing schedule rule org=%s, ID=%d: %w",
						engine.Organization.ID, rule.ID, err)
			}

			nprocessed++
			resultState, ignoredError := determineDeploymentRequestStateFromOutcome(success, ruleset.mode, isLastRule(nAlreadyProcessed, nprocessed, totalRules))
			engine.Db.Logger.Info(context.Background(),
				"Processed schedule rule: org=%s, ID=%d, success=%t, ignoredError=%t, resultState=%s",
				engine.Organization.ID, rule.ID, success, ignoredError, resultState)
			if !outcomeAlreadyRecorded {
				event, err := engine.createRuleProcessedEvent(resultState, ignoredError)
				if err != nil {
					return deploymentrequeststate.Rejected, nprocessed,
						fmt.Errorf("Error recording deployment request event: %w", err)
				}
				err = engine.createScheduleRuleOutcome(rule, event, success)
				if err != nil {
					return deploymentrequeststate.Rejected, nprocessed,
						fmt.Errorf("Error recording schedule approval rule outcome: %w", err)
				}
			}
			if resultState.IsFinal() {
				return resultState, nprocessed, nil
			}
		}
	}

	return determineDeploymentRequestStateAfterProcessingRules(nAlreadyProcessed, nprocessed, totalRules),
		nprocessed, nil
}

func determineDeploymentRequestStateAfterProcessingRules(nAlreadyProcessed uint, nprocessed uint, totalRules uint) deploymentrequeststate.State {
	if isLastRule(nAlreadyProcessed, nprocessed, totalRules) {
		return deploymentrequeststate.Approved
	}
	return deploymentrequeststate.InProgress
}

func (engine Engine) processScheduleRule(rule dbmodels.ScheduleApprovalRule, previousOutcomes map[uint64]bool) (bool, bool, error) {
	success, exists := previousOutcomes[rule.ID]
	if exists {
		return success, true, nil
	}

	success, err := timeIsWithinSchedule(engine.ReleaseBackgroundJob.DeploymentRequest.CreatedAt, rule)
	return success, false, err
}

func (engine Engine) createScheduleRuleOutcome(rule dbmodels.ScheduleApprovalRule, event dbmodels.DeploymentRequestRuleProcessedEvent, success bool) error {
	outcome := dbmodels.ScheduleApprovalRuleOutcome{
		ApprovalRuleOutcome: dbmodels.ApprovalRuleOutcome{
			BaseModel: dbmodels.BaseModel{
				OrganizationID: engine.Organization.ID,
			},
			DeploymentRequestRuleProcessedEventID: event.DeploymentRequestEvent.ID,
			Success:                               success,
		},
		ScheduleApprovalRuleID: rule.ApprovalRule.ID,
	}
	tx := engine.Db.Create(&outcome)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

func indexScheduleRuleOutcomes(outcomes []dbmodels.ScheduleApprovalRuleOutcome) map[uint64]bool {
	result := make(map[uint64]bool)
	for _, outcome := range outcomes {
		result[outcome.ApprovalRuleOutcome.ID] = outcome.Success
	}
	return result
}

func timeIsWithinSchedule(deploymentTime time.Time, rule dbmodels.ScheduleApprovalRule) (bool, error) {
	if rule.BeginTime.Valid {
		if !rule.EndTime.Valid {
			panic(fmt.Sprintf("ScheduleApprovalRule %d: BeginTime non-null, but EndTime null", rule.ApprovalRule.ID))
		}

		parsedBeginTime, err := parseScheduleTime(deploymentTime, rule.BeginTime.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing begin time '%s': %w", rule.BeginTime.String, err)
		}

		parsedEndTime, err := parseScheduleTime(deploymentTime, rule.EndTime.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing end time '%s': %w", rule.EndTime.String, err)
		}

		if deploymentTime.Before(parsedBeginTime) || deploymentTime.After(parsedEndTime) {
			return false, nil
		}
	}

	if rule.DaysOfWeek.Valid {
		parsedWeekDays, err := parseScheduleWeekDays(rule.DaysOfWeek.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing days of week '%s': %w", rule.DaysOfWeek.String, err)
		}

		if !weekDaysListContains(parsedWeekDays, deploymentTime.Weekday()) {
			return false, nil
		}
	}

	if rule.DaysOfMonth.Valid {
		parsedMonthDays, err := parseScheduleMonthDays(rule.DaysOfMonth.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing days of month '%s': %w", rule.DaysOfMonth.String, err)
		}

		if !intListContains(parsedMonthDays, deploymentTime.Day()) {
			return false, nil
		}
	}

	if rule.MonthsOfYear.Valid {
		parsedMonths, err := parseScheduleMonths(rule.MonthsOfYear.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing months '%s': %w", rule.MonthsOfYear.String, err)
		}

		if !monthListContains(parsedMonths, deploymentTime.Month()) {
			return false, nil
		}
	}

	return true, nil
}

// parseScheduleTime parses a ScheduleApprovalRule time string. It returns a `time.Time`
// whose date is equal to `date`, but whose time equals that of the time string.
//
// `str` has the format of `HH:MM[:SS]`.
//
// Example:
//
// ~~~go
// parseScheduleTime(time.Date(2021, 2, 19, 0, 0, 0), "12:32")
// // => 2021-02-19 12:32
// ~~~
func parseScheduleTime(date time.Time, str string) (time.Time, error) {
	components := strings.SplitN(str, ":", 3)
	if len(components) < 2 {
		return time.Time{}, errors.New("Invalid time format (HH:MM[:SS] expected)")
	}

	var hour, minute, second int64
	var err error

	hour, err = strconv.ParseInt(components[0], 10, 8)
	if err != nil {
		return time.Time{}, fmt.Errorf("Error parsing hour component: %w", err)
	}
	if hour < 0 || hour > 24 {
		return time.Time{}, fmt.Errorf("Error parsing hour component: %d is not a valid value", hour)
	}

	minute, err = strconv.ParseInt(components[1], 10, 8)
	if err != nil {
		return time.Time{}, fmt.Errorf("Error parsing minute component: %w", err)
	}
	if minute < 0 || minute > 60 {
		return time.Time{}, fmt.Errorf("Error parsing minute component: %d is not a valid value", minute)
	}

	if len(components) == 3 {
		second, err = strconv.ParseInt(components[2], 10, 8)
		if err != nil {
			return time.Time{}, fmt.Errorf("Error parsing second component: %w", err)
		}
		if second < 0 || second > 60 {
			return time.Time{}, fmt.Errorf("Error parsing second component: %d is not a valid value", second)
		}
	} else {
		second = 0
	}

	result := time.Date(date.Year(), date.Month(), date.Day(), int(hour), int(minute), int(second), 0, date.Location())
	return result, nil
}

func parseScheduleWeekDays(str string) ([]time.Weekday, error) {
	return nil, nil // TODO
}

func parseScheduleMonthDays(str string) ([]int, error) {
	return nil, nil // TODO
}

func parseScheduleMonths(str string) ([]time.Month, error) {
	return nil, nil // TODO
}

func weekDaysListContains(list []time.Weekday, weekDay time.Weekday) bool {
	for _, elem := range list {
		if weekDay == elem {
			return true
		}
	}
	return false
}

func intListContains(list []int, val int) bool {
	for _, elem := range list {
		if val == elem {
			return true
		}
	}
	return false
}

func monthListContains(list []time.Month, month time.Month) bool {
	for _, elem := range list {
		if month == elem {
			return true
		}
	}
	return false
}

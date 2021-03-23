package approvalrulesengine

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/releasestate"
	"gorm.io/gorm"
)

func (engine Engine) loadScheduleRules(db *gorm.DB, majorVersionIndex map[uint64]*ruleset, versionKeys []dbmodels.ApprovalRulesetVersionKey) (uint, error) {
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
	outcomes, err := dbmodels.FindAllScheduleApprovalRuleOutcomes(engine.Db, engine.Organization.ID, engine.ReleaseBackgroundJob.ReleaseID)
	if err != nil {
		return nil, err
	}

	return indexScheduleRuleOutcomes(outcomes), nil
}

func (engine Engine) processScheduleRules(rulesets []ruleset, previousOutcomes map[uint64]bool, nAlreadyProcessed uint, totalRules uint) (releasestate.State, uint, error) {
	var nprocessed uint = 0

	for _, ruleset := range rulesets {
		for _, rule := range ruleset.scheduleRules {
			success, outcomeAlreadyRecorded, err := engine.processScheduleRule(rule, previousOutcomes)
			if err != nil {
				return releasestate.Rejected, nprocessed,
					maybeFormatRuleProcessingError(err, "Error processing schedule rule org=%s, ID=%d: %w",
						engine.Organization.ID, rule.ID, err)
			}

			nprocessed++
			resultState, ignoredError := determineReleaseStateFromOutcome(success, ruleset.mode, isLastRule(nAlreadyProcessed, nprocessed, totalRules))
			engine.Db.Logger.Info(context.Background(),
				"Processed schedule rule: org=%s, ID=%d, success=%t, ignoredError=%t, resultState=%s",
				engine.Organization.ID, rule.ID, success, ignoredError, resultState)
			if !outcomeAlreadyRecorded {
				event, err := engine.createRuleProcessedEvent(resultState, ignoredError)
				if err != nil {
					return releasestate.Rejected, nprocessed,
						fmt.Errorf("Error recording release event: %w", err)
				}
				err = engine.createScheduleRuleOutcome(rule, event, success)
				if err != nil {
					return releasestate.Rejected, nprocessed,
						fmt.Errorf("Error recording schedule approval rule outcome: %w", err)
				}
			}
			if resultState.IsFinal() {
				return resultState, nprocessed, nil
			}
		}
	}

	return determineReleaseStateAfterProcessingRules(nAlreadyProcessed, nprocessed, totalRules),
		nprocessed, nil
}

func determineReleaseStateAfterProcessingRules(nAlreadyProcessed uint, nprocessed uint, totalRules uint) releasestate.State {
	if isLastRule(nAlreadyProcessed, nprocessed, totalRules) {
		return releasestate.Approved
	}
	return releasestate.InProgress
}

func (engine Engine) processScheduleRule(rule dbmodels.ScheduleApprovalRule, previousOutcomes map[uint64]bool) (bool, bool, error) {
	success, exists := previousOutcomes[rule.ID]
	if exists {
		return success, true, nil
	}

	// TODO: if there's an error, reject the release because the rules have errors
	success, err := timeIsWithinSchedule(engine.ReleaseBackgroundJob.Release.CreatedAt, rule)
	return success, false, err
}

func (engine Engine) createScheduleRuleOutcome(rule dbmodels.ScheduleApprovalRule, event dbmodels.ReleaseRuleProcessedEvent, success bool) error {
	outcome := dbmodels.ScheduleApprovalRuleOutcome{
		ApprovalRuleOutcome: dbmodels.ApprovalRuleOutcome{
			BaseModel: dbmodels.BaseModel{
				OrganizationID: engine.Organization.ID,
			},
			ReleaseRuleProcessedEventID: event.ReleaseEvent.ID,
			Success:                     success,
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

		if !parsedWeekDays[deploymentTime.Weekday()] {
			return false, nil
		}
	}

	if rule.DaysOfMonth.Valid {
		parsedMonthDays, err := parseScheduleMonthDays(rule.DaysOfMonth.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing days of month '%s': %w", rule.DaysOfMonth.String, err)
		}

		if !parsedMonthDays[deploymentTime.Day()] {
			return false, nil
		}
	}

	if rule.MonthsOfYear.Valid {
		parsedMonths, err := parseScheduleMonths(rule.MonthsOfYear.String)
		if err != nil {
			return false, fmt.Errorf("Error parsing months '%s': %w", rule.MonthsOfYear.String, err)
		}

		if !parsedMonths[deploymentTime.Month()] {
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

func parseScheduleWeekDays(str string) (map[time.Weekday]bool, error) {
	result := make(map[time.Weekday]bool)
	for _, day := range strings.Split(str, " ") {
		switch strings.ToLower(day) {
		case "1", "mon", "monday":
			result[time.Monday] = true
		case "2", "tue", "tuesday":
			result[time.Tuesday] = true
		case "3", "wed", "wednesday":
			result[time.Wednesday] = true
		case "4", "thu", "thursday":
			result[time.Thursday] = true
		case "5", "fri", "friday":
			result[time.Friday] = true
		case "6", "sat", "saturday":
			result[time.Saturday] = true
		case "0", "7", "sun", "sunday":
			result[time.Sunday] = true
		case "":
			continue
		default:
			return nil, fmt.Errorf("'%s' is not a recognized weekday", day)
		}
	}
	return result, nil
}

func parseScheduleMonthDays(str string) (map[int]bool, error) {
	result := make(map[int]bool)
	for _, day := range strings.Split(str, " ") {
		if len(day) == 0 {
			continue
		}

		dayInt, err := strconv.Atoi(day)
		if err != nil {
			return nil, fmt.Errorf("Error parsing month day '%s': %w", day, err)
		}

		if dayInt < 0 || dayInt > 31 {
			return nil, fmt.Errorf("Month day '%s' is not a valid day", day)
		}

		result[int(dayInt)] = true
	}
	return result, nil
}

func parseScheduleMonths(str string) (map[time.Month]bool, error) {
	result := make(map[time.Month]bool)
	for _, month := range strings.Split(str, " ") {
		switch strings.ToLower(month) {
		case "1", "jan", "january":
			result[time.January] = true
		case "2", "feb", "february":
			result[time.February] = true
		case "3", "mar", "march":
			result[time.March] = true
		case "4", "apr", "april":
			result[time.April] = true
		case "5", "may":
			result[time.May] = true
		case "6", "jun", "june":
			result[time.June] = true
		case "7", "jul", "july":
			result[time.July] = true
		case "8", "aug", "august":
			result[time.August] = true
		case "9", "sep", "september":
			result[time.September] = true
		case "10", "oct", "october":
			result[time.October] = true
		case "11", "nov", "november":
			result[time.November] = true
		case "12", "dec", "december":
			result[time.December] = true
		case "":
			continue
		default:
			return nil, fmt.Errorf("'%s' is not a recognized month", month)
		}
	}
	return result, nil
}

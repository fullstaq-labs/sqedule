package approvalrulesengine

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/fullstaq-labs/sqedule/server/dbmodels"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/server/dbmodels/releasestate"
	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Test processScheduleRules()

type ProcessScheduleRulesTestContext struct {
	db                *gorm.DB
	org               dbmodels.Organization
	app               dbmodels.Application
	release           dbmodels.Release
	permissiveBinding dbmodels.ApplicationApprovalRulesetBinding
	enforcingBinding  dbmodels.ApplicationApprovalRulesetBinding
	job               dbmodels.ReleaseBackgroundJob
	engine            Engine
	rulesets          []ruleset
	permissiveRuleset *ruleset
	enforcingRuleset  *ruleset
}

func setupProcessScheduleRulesTest() (ProcessScheduleRulesTestContext, error) {
	var ctx ProcessScheduleRulesTestContext
	var err error

	ctx.db, err = dbutils.SetupTestDatabase()
	if err != nil {
		return ProcessScheduleRulesTestContext{}, err
	}

	err = ctx.db.Transaction(func(tx *gorm.DB) error {
		ctx.org, err = dbmodels.CreateMockOrganization(tx)
		if err != nil {
			return err
		}

		ctx.app, err = dbmodels.CreateMockApplicationWith1Version(tx, ctx.org, nil, nil)
		if err != nil {
			return err
		}

		ctx.release, err = dbmodels.CreateMockReleaseWithInProgressState(tx, ctx.org, ctx.app, func(release *dbmodels.Release) {
			release.CreatedAt = time.Date(2020, time.March, 3, 12, 0, 0, 0, time.Now().Local().Location())
		})
		if err != nil {
			return err
		}

		ctx.permissiveBinding, ctx.enforcingBinding, err = dbmodels.CreateMockApplicationApprovalRulesetsAndBindingsWith2Modes1Version(tx, ctx.org, ctx.app)
		if err != nil {
			return err
		}

		ctx.job, err = dbmodels.CreateMockReleaseBackgroundJob(tx, ctx.org, ctx.app, ctx.release)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return ProcessScheduleRulesTestContext{}, err
	}

	ctx.engine = Engine{
		Db:                   ctx.db,
		Organization:         ctx.org,
		ReleaseBackgroundJob: ctx.job,
	}
	ctx.rulesets = []ruleset{
		{
			ApprovalRulesetContents: dbmodels.ApprovalRulesetContents{
				ManualApprovalRules:   []dbmodels.ManualApprovalRule{},
				ScheduleApprovalRules: []dbmodels.ScheduleApprovalRule{},
				HTTPApiApprovalRules:  []dbmodels.HTTPApiApprovalRule{},
			},
			mode: approvalrulesetbindingmode.Permissive,
		},
		{
			ApprovalRulesetContents: dbmodels.ApprovalRulesetContents{
				ManualApprovalRules:   []dbmodels.ManualApprovalRule{},
				ScheduleApprovalRules: []dbmodels.ScheduleApprovalRule{},
				HTTPApiApprovalRules:  []dbmodels.HTTPApiApprovalRule{},
			},
			mode: approvalrulesetbindingmode.Enforcing,
		},
	}
	ctx.permissiveRuleset = &ctx.rulesets[0]
	ctx.enforcingRuleset = &ctx.rulesets[1]
	return ctx, nil
}

func TestProcessScheduleRulesSuccess(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	rule, err := dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.db, ctx.org,
		ctx.enforcingBinding.ApprovalRuleset.LatestMajorVersion.ID,
		*ctx.enforcingBinding.ApprovalRuleset.LatestMinorVersion,
		nil)
	if !assert.NoError(t, err) {
		return
	}
	ctx.enforcingRuleset.ScheduleApprovalRules = append(ctx.enforcingRuleset.ScheduleApprovalRules, rule)

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.ReleaseRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.ReleaseRuleProcessedEvent{}).Count(&numProcessedEvents).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Model(&dbmodels.ScheduleApprovalRuleOutcome{}).Count(&numOutcomes).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&event).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&outcome).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, int64(1), numProcessedEvents)
	assert.Equal(t, int64(1), numOutcomes)
	assert.Equal(t, releasestate.Approved, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.True(t, outcome.Success)
	assert.Equal(t, releasestate.Approved, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

func TestProcessScheduleRulesError(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	rule, err := dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.db, ctx.org,
		ctx.enforcingBinding.ApprovalRuleset.LatestMajorVersion.ID,
		*ctx.enforcingBinding.ApprovalRuleset.LatestMinorVersion,
		func(r *dbmodels.ScheduleApprovalRule) {
			r.BeginTime = sql.NullString{String: "0:00:00", Valid: true}
			r.EndTime = sql.NullString{String: "0:00:01", Valid: true}
		})
	if !assert.NoError(t, err) {
		return
	}
	ctx.enforcingRuleset.ScheduleApprovalRules = append(ctx.enforcingRuleset.ScheduleApprovalRules, rule)

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.ReleaseRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.ReleaseRuleProcessedEvent{}).Count(&numProcessedEvents).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Model(&dbmodels.ScheduleApprovalRuleOutcome{}).Count(&numOutcomes).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&event).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&outcome).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, int64(1), numProcessedEvents)
	assert.Equal(t, int64(1), numOutcomes)
	assert.Equal(t, releasestate.Rejected, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.False(t, outcome.Success)
	assert.Equal(t, releasestate.Rejected, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

func TestProcessScheduleRulesPermissiveMode(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	rule, err := dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.db, ctx.org,
		ctx.permissiveBinding.ApprovalRuleset.LatestMajorVersion.ID,
		*ctx.permissiveBinding.ApprovalRuleset.LatestMinorVersion,
		func(r *dbmodels.ScheduleApprovalRule) {
			r.BeginTime = sql.NullString{String: "0:00:00", Valid: true}
			r.EndTime = sql.NullString{String: "0:00:01", Valid: true}
		})
	if !assert.NoError(t, err) {
		return
	}
	ctx.permissiveRuleset.ScheduleApprovalRules = append(ctx.permissiveRuleset.ScheduleApprovalRules, rule)

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.ReleaseRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.ReleaseRuleProcessedEvent{}).Count(&numProcessedEvents).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Model(&dbmodels.ScheduleApprovalRuleOutcome{}).Count(&numOutcomes).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&event).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&outcome).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, int64(1), numProcessedEvents)
	assert.Equal(t, int64(1), numOutcomes)
	assert.Equal(t, releasestate.Approved, event.ResultState)
	assert.True(t, event.IgnoredError)
	assert.False(t, outcome.Success)
	assert.Equal(t, releasestate.Approved, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

func TestProcessScheduleRulesEmptyRuleset(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 0)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, releasestate.Approved, resultState)
	assert.Equal(t, uint(0), nprocessed)
}

func TestProcessScheduleRulesRerunSuccess(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	rule, err := dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.db, ctx.org,
		ctx.enforcingBinding.ApprovalRuleset.LatestMajorVersion.ID,
		*ctx.enforcingBinding.ApprovalRuleset.LatestMinorVersion,
		nil)
	if !assert.NoError(t, err) {
		return
	}
	ctx.enforcingRuleset.ScheduleApprovalRules = append(ctx.enforcingRuleset.ScheduleApprovalRules, rule)

	_, _, err = ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}
	outcomes, err := dbmodels.FindAllScheduleApprovalRuleOutcomes(ctx.db, ctx.org.ID, ctx.release.ID)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, 1, len(outcomes)) {
		return
	}

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, indexScheduleRuleOutcomes(outcomes), 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.ReleaseRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.ReleaseRuleProcessedEvent{}).Count(&numProcessedEvents).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Model(&dbmodels.ScheduleApprovalRuleOutcome{}).Count(&numOutcomes).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&event).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&outcome).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, int64(1), numProcessedEvents)
	assert.Equal(t, int64(1), numOutcomes)
	assert.Equal(t, releasestate.Approved, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.True(t, outcome.Success)
	assert.Equal(t, releasestate.Approved, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

func TestProcessScheduleRulesRerunFail(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	rule, err := dbmodels.CreateMockScheduleApprovalRuleWholeDay(ctx.db, ctx.org,
		ctx.enforcingBinding.ApprovalRuleset.LatestMajorVersion.ID,
		*ctx.enforcingBinding.ApprovalRuleset.LatestMinorVersion,
		func(r *dbmodels.ScheduleApprovalRule) {
			r.BeginTime = sql.NullString{String: "1:00", Valid: true}
			r.EndTime = sql.NullString{String: "1:01", Valid: true}
		})
	if !assert.NoError(t, err) {
		return
	}
	ctx.enforcingRuleset.ScheduleApprovalRules = append(ctx.enforcingRuleset.ScheduleApprovalRules, rule)

	_, _, err = ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}
	outcomes, err := dbmodels.FindAllScheduleApprovalRuleOutcomes(ctx.db, ctx.org.ID, ctx.release.ID)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Equal(t, 1, len(outcomes)) {
		return
	}

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, indexScheduleRuleOutcomes(outcomes), 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.ReleaseRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.ReleaseRuleProcessedEvent{}).Count(&numProcessedEvents).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Model(&dbmodels.ScheduleApprovalRuleOutcome{}).Count(&numOutcomes).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&event).Error
	if !assert.NoError(t, err) {
		return
	}
	err = ctx.db.Take(&outcome).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, int64(1), numProcessedEvents)
	assert.Equal(t, int64(1), numOutcomes)
	assert.Equal(t, releasestate.Rejected, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.False(t, outcome.Success)
	assert.Equal(t, releasestate.Rejected, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

// Test parseScheduleTime()

func TestParseScheduleTimeTooFewComponents(t *testing.T) {
	inputs := []string{"", "1"}
	for _, input := range inputs {
		_, err := parseScheduleTime(time.Now(), input)
		if assert.Error(t, err, "Input=%s", input) {
			assert.Regexp(t, "Invalid time format", err.Error(), "Input=%s", input)
		}
	}
}

func TestParseScheduleTimeInvalidValues(t *testing.T) {
	type Input struct {
		Value     string
		Component string
	}

	inputs := []Input{
		{Value: ":", Component: "hour"},
		{Value: "a:", Component: "hour"},
		{Value: "-1:", Component: "hour"},
		{Value: "25:", Component: "hour"},

		{Value: "1:", Component: "minute"},
		{Value: "1:b", Component: "minute"},
		{Value: "1:-1", Component: "minute"},
		{Value: "1:61", Component: "minute"},

		{Value: "1:30:", Component: "second"},
		{Value: "1:30:c", Component: "second"},
		{Value: "1:30:-1", Component: "second"},
		{Value: "1:30:61", Component: "second"},
	}

	for _, input := range inputs {
		_, err := parseScheduleTime(time.Now(), input.Value)
		if assert.Error(t, err, "Input=%#v", input) {
			assert.Regexp(t, "Error parsing "+input.Component+" component",
				err.Error(), "Input=%#v", input)
		}
	}
}

func TestParseScheduleTimeValidValues(t *testing.T) {
	var err error
	var parsed time.Time
	now := time.Now()

	parsed, err = parseScheduleTime(now, "1:20")
	if assert.NoError(t, err) {
		assert.Equal(t, parsed.Hour(), 1)
		assert.Equal(t, parsed.Minute(), 20)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "01:20")
	if assert.NoError(t, err) {
		assert.Equal(t, parsed.Hour(), 1)
		assert.Equal(t, parsed.Minute(), 20)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "16:5")
	if assert.NoError(t, err) {
		assert.Equal(t, parsed.Hour(), 16)
		assert.Equal(t, parsed.Minute(), 5)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "16:05")
	if assert.NoError(t, err) {
		assert.Equal(t, parsed.Hour(), 16)
		assert.Equal(t, parsed.Minute(), 5)
		assert.Equal(t, parsed.Second(), 0)
	}

	parsed, err = parseScheduleTime(now, "8:47:1")
	if assert.NoError(t, err) {
		assert.Equal(t, parsed.Hour(), 8)
		assert.Equal(t, parsed.Minute(), 47)
		assert.Equal(t, parsed.Second(), 1)
	}

	parsed, err = parseScheduleTime(now, "8:47:01")
	if assert.NoError(t, err) {
		assert.Equal(t, parsed.Hour(), 8)
		assert.Equal(t, parsed.Minute(), 47)
		assert.Equal(t, parsed.Second(), 1)
	}
}

// Test parseScheduleWeekDays()

func TestParseScheduleWeekDaysEmpty(t *testing.T) {
	var parsed map[time.Weekday]bool
	var err error

	parsed, err = parseScheduleWeekDays("")
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(parsed))
	}
}

func TestParseScheduleWeekDaysFull(t *testing.T) {
	var parsed map[time.Weekday]bool
	var err error

	parsed, err = parseScheduleWeekDays("monday")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.Monday])
	}

	parsed, err = parseScheduleWeekDays("monday tuesday wednesday")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[time.Monday])
		assert.True(t, parsed[time.Tuesday])
		assert.True(t, parsed[time.Wednesday])
	}

	parsed, err = parseScheduleWeekDays("sunday thursday friday saturday")
	if assert.NoError(t, err) {
		assert.Equal(t, 4, len(parsed))
		assert.True(t, parsed[time.Thursday])
		assert.True(t, parsed[time.Friday])
		assert.True(t, parsed[time.Saturday])
		assert.True(t, parsed[time.Sunday])
	}
}

func TestParseScheduleWeekDaysAbbrev(t *testing.T) {
	var parsed map[time.Weekday]bool
	var err error

	parsed, err = parseScheduleWeekDays("mon")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.Monday])
	}

	parsed, err = parseScheduleWeekDays("mon tue wed")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[time.Monday])
		assert.True(t, parsed[time.Tuesday])
		assert.True(t, parsed[time.Wednesday])
	}

	parsed, err = parseScheduleWeekDays("sun thu fri sat")
	if assert.NoError(t, err) {
		assert.Equal(t, 4, len(parsed))
		assert.True(t, parsed[time.Thursday])
		assert.True(t, parsed[time.Friday])
		assert.True(t, parsed[time.Saturday])
		assert.True(t, parsed[time.Sunday])
	}
}

func TestParseScheduleWeekDaysNumbers(t *testing.T) {
	var parsed map[time.Weekday]bool
	var err error

	parsed, err = parseScheduleWeekDays("1")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.Monday])
	}

	parsed, err = parseScheduleWeekDays("1 2 3")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[time.Monday])
		assert.True(t, parsed[time.Tuesday])
		assert.True(t, parsed[time.Wednesday])
	}

	parsed, err = parseScheduleWeekDays("7 4 5 6")
	if assert.NoError(t, err) {
		assert.Equal(t, 4, len(parsed))
		assert.True(t, parsed[time.Thursday])
		assert.True(t, parsed[time.Friday])
		assert.True(t, parsed[time.Saturday])
		assert.True(t, parsed[time.Sunday])
	}

	parsed, err = parseScheduleWeekDays("0")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.Sunday])
	}
}

func TestParseScheduleWeekDaysExcessiveSpaces(t *testing.T) {
	var parsed map[time.Weekday]bool
	var err error

	parsed, err = parseScheduleWeekDays("  mon  wed    ")
	if assert.NoError(t, err) {
		assert.Equal(t, 2, len(parsed))
		assert.True(t, parsed[time.Monday])
		assert.True(t, parsed[time.Wednesday])
	}
}

func TestParseScheduleWeekDaysCaseInsensitive(t *testing.T) {
	var parsed map[time.Weekday]bool
	var err error

	inputs := []string{
		"Mon Tue Wed Thu Fri Sat Sun",
		"MON TUE WED THU FRI SAT SUN",
		"monDay tuesDay wednesDay thursDay friDay saturDay sunDay",
		"MONDAY TUESDAY WEDNESDAY THURSDAY FRIDAY SATURDAY SUNDAY",
	}

	for _, input := range inputs {
		parsed, err = parseScheduleWeekDays(input)
		if assert.NoError(t, err) {
			assert.Equal(t, 7, len(parsed))
			assert.True(t, parsed[time.Monday], "Input=%s", input)
			assert.True(t, parsed[time.Tuesday], "Input=%s", input)
			assert.True(t, parsed[time.Wednesday], "Input=%s", input)
			assert.True(t, parsed[time.Thursday], "Input=%s", input)
			assert.True(t, parsed[time.Friday], "Input=%s", input)
			assert.True(t, parsed[time.Saturday], "Input=%s", input)
			assert.True(t, parsed[time.Sunday], "Input=%s", input)
		}
	}
}

func TestParseScheduleWeekDaysUnknownInput(t *testing.T) {
	var err error

	_, err = parseScheduleWeekDays("today")
	assert.Error(t, err)

	_, err = parseScheduleWeekDays("mon:tue")
	assert.Error(t, err)
}

// Test parseScheduleMonthDays()

func TestParseScheduleMonthDaysEmpty(t *testing.T) {
	var parsed map[int]bool
	var err error

	parsed, err = parseScheduleMonthDays("")
	if assert.NoError(t, err) {
		assert.Equal(t, 0, len(parsed))
	}
}

func TestParseScheduleMonthDays(t *testing.T) {
	var parsed map[int]bool
	var err error

	parsed, err = parseScheduleMonthDays("1")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[1])
	}

	parsed, err = parseScheduleMonthDays("1 2 3")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[1])
		assert.True(t, parsed[2])
		assert.True(t, parsed[3])
	}

	var input = ""
	for i := 1; i <= 31; i++ {
		input += fmt.Sprintf("%d", i) + " "
	}
	parsed, err = parseScheduleMonthDays(input)
	if assert.NoError(t, err) {
		assert.Equal(t, 31, len(parsed))
		for i := 1; i <= 31; i++ {
			assert.True(t, parsed[i])
		}
	}
}

func TestParseScheduleMonthDaysExcessiveSpaces(t *testing.T) {
	var parsed map[int]bool
	var err error

	parsed, err = parseScheduleMonthDays("  1  15  30    ")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[1])
		assert.True(t, parsed[15])
		assert.True(t, parsed[30])
	}
}

func TestParseScheduleMonthDaysInvalidInput(t *testing.T) {
	var err error

	_, err = parseScheduleMonthDays("aa")
	assert.Error(t, err)

	_, err = parseScheduleMonthDays("1:2")
	assert.Error(t, err)

	_, err = parseScheduleMonthDays("32")
	assert.Error(t, err)

	_, err = parseScheduleMonthDays("-1")
	assert.Error(t, err)
}

// Test parseScheduleMonths()

func TestParseScheduleMonthsFull(t *testing.T) {
	var parsed map[time.Month]bool
	var err error

	parsed, err = parseScheduleMonths("january")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.January])
	}

	parsed, err = parseScheduleMonths("january february march")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[time.January])
		assert.True(t, parsed[time.February])
		assert.True(t, parsed[time.March])
	}

	parsed, err = parseScheduleMonths("december april may june july august september october november")
	if assert.NoError(t, err) {
		assert.Equal(t, 9, len(parsed))
		assert.True(t, parsed[time.April])
		assert.True(t, parsed[time.May])
		assert.True(t, parsed[time.June])
		assert.True(t, parsed[time.July])
		assert.True(t, parsed[time.August])
		assert.True(t, parsed[time.September])
		assert.True(t, parsed[time.October])
		assert.True(t, parsed[time.November])
		assert.True(t, parsed[time.December])
	}
}

func TestParseScheduleMonthsAbbrev(t *testing.T) {
	var parsed map[time.Month]bool
	var err error

	parsed, err = parseScheduleMonths("jan")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.January])
	}

	parsed, err = parseScheduleMonths("jan feb mar")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[time.January])
		assert.True(t, parsed[time.February])
		assert.True(t, parsed[time.March])
	}

	parsed, err = parseScheduleMonths("dec apr may jun jul aug sep oct nov")
	if assert.NoError(t, err) {
		assert.Equal(t, 9, len(parsed))
		assert.True(t, parsed[time.April])
		assert.True(t, parsed[time.May])
		assert.True(t, parsed[time.June])
		assert.True(t, parsed[time.July])
		assert.True(t, parsed[time.August])
		assert.True(t, parsed[time.September])
		assert.True(t, parsed[time.October])
		assert.True(t, parsed[time.November])
		assert.True(t, parsed[time.December])
	}
}

func TestParseScheduleMonthsNumbers(t *testing.T) {
	var parsed map[time.Month]bool
	var err error

	parsed, err = parseScheduleMonths("1")
	if assert.NoError(t, err) {
		assert.Equal(t, 1, len(parsed))
		assert.True(t, parsed[time.January])
	}

	parsed, err = parseScheduleMonths("1 2 3")
	if assert.NoError(t, err) {
		assert.Equal(t, 3, len(parsed))
		assert.True(t, parsed[time.January])
		assert.True(t, parsed[time.February])
		assert.True(t, parsed[time.March])
	}

	parsed, err = parseScheduleMonths("7 4 5 6 7 8 9 10 11 12")
	if assert.NoError(t, err) {
		assert.Equal(t, 9, len(parsed))
		assert.True(t, parsed[time.April])
		assert.True(t, parsed[time.May])
		assert.True(t, parsed[time.June])
		assert.True(t, parsed[time.July])
		assert.True(t, parsed[time.August])
		assert.True(t, parsed[time.September])
		assert.True(t, parsed[time.October])
		assert.True(t, parsed[time.November])
		assert.True(t, parsed[time.December])
	}
}

func TestParseScheduleMonthsExcessiveSpaces(t *testing.T) {
	var parsed map[time.Month]bool
	var err error

	parsed, err = parseScheduleMonths("  jan  feb    ")
	if assert.NoError(t, err) {
		assert.Equal(t, 2, len(parsed))
		assert.True(t, parsed[time.January])
		assert.True(t, parsed[time.February])
	}
}

func TestParseScheduleMonthsCaseInsensitive(t *testing.T) {
	var parsed map[time.Month]bool
	var err error

	inputs := []string{
		"Jan Feb Mar Apr May Jun Jul Aug Sep Oct Nov Dec",
		"JAN FEB MAR APR MAY JUN JUL AUG SEP OCT NOV DEC",
		"januAry februAry marCh apRil maY juNe juLy auGust sepTember ocTober noVember deCember",
		"JANUARY FEBRUARY MARCH APRIL MAY JUNE JULY AUGUST SEPTEMBER OCTOBER NOVEMBER DECEMBER",
	}

	for _, input := range inputs {
		parsed, err = parseScheduleMonths(input)
		if assert.NoError(t, err) {
			assert.Equal(t, 12, len(parsed))
			assert.True(t, parsed[time.January], "Input=%s", input)
			assert.True(t, parsed[time.February], "Input=%s", input)
			assert.True(t, parsed[time.March], "Input=%s", input)
			assert.True(t, parsed[time.April], "Input=%s", input)
			assert.True(t, parsed[time.May], "Input=%s", input)
			assert.True(t, parsed[time.June], "Input=%s", input)
			assert.True(t, parsed[time.July], "Input=%s", input)
			assert.True(t, parsed[time.August], "Input=%s", input)
			assert.True(t, parsed[time.September], "Input=%s", input)
			assert.True(t, parsed[time.October], "Input=%s", input)
			assert.True(t, parsed[time.November], "Input=%s", input)
			assert.True(t, parsed[time.December], "Input=%s", input)
		}
	}
}

func TestParseScheduleMonthsUnknownInput(t *testing.T) {
	var err error

	_, err = parseScheduleMonths("today")
	assert.Error(t, err)

	_, err = parseScheduleMonths("jan:feb")
	assert.Error(t, err)

	_, err = parseScheduleMonths("-1")
	assert.Error(t, err)

	_, err = parseScheduleMonths("13")
	assert.Error(t, err)
}

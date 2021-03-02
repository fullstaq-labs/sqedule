package approvalrulesengine

import (
	"database/sql"
	"testing"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels"
	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/dbmodels/deploymentrequeststate"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// Test processScheduleRules()

type ProcessScheduleRulesTestContext struct {
	db                *gorm.DB
	org               dbmodels.Organization
	app               dbmodels.Application
	deploymentRequest dbmodels.DeploymentRequest
	permissiveBinding dbmodels.ApprovalRulesetBinding
	enforcingBinding  dbmodels.ApprovalRulesetBinding
	job               dbmodels.ReleaseBackgroundJob
	engine            Engine
	rulesets          []ruleset
	permissiveRuleset *ruleset
	enforcingRuleset  *ruleset
	baseApprovalRule  dbmodels.ApprovalRule
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

		ctx.app, err = dbmodels.CreateMockApplicationWithOneVersion(tx, ctx.org)
		if err != nil {
			return err
		}

		ctx.deploymentRequest, err = dbmodels.CreateMockDeploymentRequestWithInProgressState(tx, ctx.org, ctx.app, func(dr *dbmodels.DeploymentRequest) {
			dr.CreatedAt = time.Date(2020, time.March, 3, 12, 0, 0, 0, time.Now().Local().Location())
		})
		if err != nil {
			return err
		}

		ctx.permissiveBinding, ctx.enforcingBinding, err = dbmodels.CreateMockApprovalRulesetsAndBindingsWith2Modes1Version(tx, ctx.org, ctx.app)
		if err != nil {
			return err
		}

		ctx.job, err = dbmodels.CreateMockReleaseBackgroundJob(tx, ctx.org, ctx.app, ctx.deploymentRequest)
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
			mode:                approvalrulesetbindingmode.Permissive,
			manualApprovalRules: []dbmodels.ManualApprovalRule{},
			scheduleRules:       []dbmodels.ScheduleApprovalRule{},
			httpAPIRules:        []dbmodels.HTTPApiApprovalRule{},
		},
		{
			mode:                approvalrulesetbindingmode.Enforcing,
			manualApprovalRules: []dbmodels.ManualApprovalRule{},
			scheduleRules:       []dbmodels.ScheduleApprovalRule{},
			httpAPIRules:        []dbmodels.HTTPApiApprovalRule{},
		},
	}
	ctx.permissiveRuleset = &ctx.rulesets[0]
	ctx.enforcingRuleset = &ctx.rulesets[1]
	ctx.baseApprovalRule = dbmodels.ApprovalRule{
		BaseModel: dbmodels.BaseModel{
			OrganizationID: ctx.org.ID,
			Organization:   ctx.org,
		},
		Enabled: true,
	}
	return ctx, nil
}

func TestProcessScheduleRulesSuccess(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	ctx.enforcingRuleset.scheduleRules = append(ctx.enforcingRuleset.scheduleRules, dbmodels.ScheduleApprovalRule{
		ApprovalRule: ctx.baseApprovalRule,
		BeginTime:    sql.NullString{String: "0:00:00", Valid: true},
		EndTime:      sql.NullString{String: "23:59:59", Valid: true},
	})

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.DeploymentRequestRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.DeploymentRequestRuleProcessedEvent{}).Count(&numProcessedEvents).Error
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
	assert.Equal(t, deploymentrequeststate.Approved, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.True(t, outcome.Success)
	assert.Equal(t, deploymentrequeststate.Approved, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

func TestProcessScheduleRulesError(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	ctx.enforcingRuleset.scheduleRules = append(ctx.enforcingRuleset.scheduleRules, dbmodels.ScheduleApprovalRule{
		ApprovalRule: ctx.baseApprovalRule,
		BeginTime:    sql.NullString{String: "0:00:00", Valid: true},
		EndTime:      sql.NullString{String: "0:00:01", Valid: true},
	})

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.DeploymentRequestRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.DeploymentRequestRuleProcessedEvent{}).Count(&numProcessedEvents).Error
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
	assert.Equal(t, deploymentrequeststate.Rejected, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.False(t, outcome.Success)
	assert.Equal(t, deploymentrequeststate.Rejected, resultState)
	assert.Equal(t, uint(1), nprocessed)
}

func TestProcessScheduleRulesPermissiveMode(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	ctx.permissiveRuleset.scheduleRules = append(ctx.permissiveRuleset.scheduleRules, dbmodels.ScheduleApprovalRule{
		ApprovalRule: ctx.baseApprovalRule,
		BeginTime:    sql.NullString{String: "0:00:00", Valid: true},
		EndTime:      sql.NullString{String: "0:00:01", Valid: true},
	})

	resultState, nprocessed, err := ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}

	var numProcessedEvents, numOutcomes int64
	var event dbmodels.DeploymentRequestRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.DeploymentRequestRuleProcessedEvent{}).Count(&numProcessedEvents).Error
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
	assert.Equal(t, deploymentrequeststate.Approved, event.ResultState)
	assert.True(t, event.IgnoredError)
	assert.False(t, outcome.Success)
	assert.Equal(t, deploymentrequeststate.Approved, resultState)
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

	assert.Equal(t, deploymentrequeststate.Approved, resultState)
	assert.Equal(t, uint(0), nprocessed)
}

func TestProcessScheduleRulesRerunSuccess(t *testing.T) {
	ctx, err := setupProcessScheduleRulesTest()
	if !assert.NoError(t, err) {
		return
	}

	ctx.enforcingRuleset.scheduleRules = append(ctx.enforcingRuleset.scheduleRules, dbmodels.ScheduleApprovalRule{
		ApprovalRule: ctx.baseApprovalRule,
		BeginTime:    sql.NullString{String: "0:00:00", Valid: true},
		EndTime:      sql.NullString{String: "23:59:59", Valid: true},
	})

	_, _, err = ctx.engine.processScheduleRules(ctx.rulesets, map[uint64]bool{}, 0, 1)
	if !assert.NoError(t, err) {
		return
	}
	outcomes, err := dbmodels.FindAllScheduleApprovalRuleOutcomes(ctx.db, ctx.org.ID, ctx.deploymentRequest.ID)
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
	var event dbmodels.DeploymentRequestRuleProcessedEvent
	var outcome dbmodels.ScheduleApprovalRuleOutcome

	err = ctx.db.Model(&dbmodels.DeploymentRequestRuleProcessedEvent{}).Count(&numProcessedEvents).Error
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
	assert.Equal(t, deploymentrequeststate.Approved, event.ResultState)
	assert.False(t, event.IgnoredError)
	assert.True(t, outcome.Success)
	assert.Equal(t, deploymentrequeststate.Approved, resultState)
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

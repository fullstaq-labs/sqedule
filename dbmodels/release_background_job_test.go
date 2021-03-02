package dbmodels

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/dbmodels/approvalrulesetbindingmode"
	"github.com/fullstaq-labs/sqedule/dbutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type CreateReleaseBackgroundJobTestContext struct {
	db                *gorm.DB
	org               Organization
	app               Application
	deploymentRequest DeploymentRequest
}

func setupCreateReleaseBackgroundJobTest() (CreateReleaseBackgroundJobTestContext, error) {
	var ctx CreateReleaseBackgroundJobTestContext
	var err error

	ctx.db, err = dbutils.SetupTestDatabase()
	if err != nil {
		return CreateReleaseBackgroundJobTestContext{}, err
	}

	err = ctx.db.Transaction(func(tx *gorm.DB) error {
		ctx.org, err = CreateMockOrganization(tx)
		if err != nil {
			return err
		}
		ctx.app, err = CreateMockApplicationWithOneVersion(tx, ctx.org)
		if err != nil {
			return err
		}
		ctx.deploymentRequest, err = CreateMockDeploymentRequestWithInProgressState(tx, ctx.org, ctx.app, nil)
		if err != nil {
			return err
		}

		return nil
	})

	return ctx, err
}

func TestCreateReleaseBackgroundJob(t *testing.T) {
	ctx, err := setupCreateReleaseBackgroundJobTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		_, numTries, err := createReleaseBackgroundJobWithDebug(tx, ctx.org, ctx.app.ID, ctx.deploymentRequest, 1)
		if !assert.NoError(t, err) {
			return nil
		}
		assert.Equal(t, uint(1), numTries)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestCreateReleaseBackgroundJob_copyBindings(t *testing.T) {
	ctx, err := setupCreateReleaseBackgroundJobTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		permissiveBinding, enforcingBinding, err := CreateMockApprovalRulesetsAndBindingsWith2Modes1Version(tx, ctx.org, ctx.app)
		if !assert.NoError(t, err) {
			return nil
		}

		_, _, err = createReleaseBackgroundJobWithDebug(tx, ctx.org, ctx.app.ID, ctx.deploymentRequest, 1)
		if !assert.NoError(t, err) {
			return nil
		}

		var count int64
		err = tx.Model(ReleaseBackgroundJobApprovalRulesetBinding{}).Count(&count).Error
		if !assert.Equal(t, int64(2), count) {
			return nil
		}

		jobBindings, err := FindAllReleaseBackgroundJobApprovalRulesetBindings(tx.Order("mode"),
			ctx.org.ID, ctx.app.ID, ctx.deploymentRequest.ID)
		if !assert.NoError(t, err) {
			return nil
		}
		if !assert.Equal(t, 2, len(jobBindings)) {
			return nil
		}
		assert.Equal(t, permissiveBinding.ApprovalRulesetID, jobBindings[0].ApprovalRulesetID)
		assert.Equal(t, enforcingBinding.ApprovalRulesetID, jobBindings[1].ApprovalRulesetID)
		assert.Equal(t, approvalrulesetbindingmode.Permissive, jobBindings[0].Mode)
		assert.Equal(t, approvalrulesetbindingmode.Enforcing, jobBindings[1].Mode)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestCreateReleaseBackgroundJob_pickRandomLockIDOnIDClash(t *testing.T) {
	ctx, err := setupCreateReleaseBackgroundJobTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		deploymentRequest2, err := CreateMockDeploymentRequestWithInProgressState(tx, ctx.org, ctx.app, nil)
		if !assert.NoError(t, err) {
			return nil
		}

		// Create a job and delete it, in order to predict what the next lock ID will be.
		job, err := CreateReleaseBackgroundJob(tx, ctx.org, ctx.app.ID, ctx.deploymentRequest)
		if !assert.NoError(t, err) {
			return nil
		}
		nextLockID := (job.LockID + 1) % ReleaseBackgroundJobMaxLockID
		err = tx.Delete(&job).Error
		if !assert.NoError(t, err) {
			return nil
		}

		// Create a job with the predicted next lock ID.
		job = ReleaseBackgroundJob{
			BaseModel: BaseModel{
				OrganizationID: ctx.org.ID,
			},
			ApplicationID:       ctx.app.ID,
			DeploymentRequestID: deploymentRequest2.ID,
			LockID:              nextLockID,
		}
		err = tx.Create(&job).Error
		if !assert.NoError(t, err) {
			return nil
		}

		// Create another job, whose autoincremented lock ID should conflict.
		job, numTries, err := createReleaseBackgroundJobWithDebug(tx, ctx.org, ctx.app.ID, ctx.deploymentRequest, 100)
		if !assert.NoError(t, err) {
			return nil
		}
		assert.Equal(t, uint(2), numTries)
		assert.NotEqual(t, (nextLockID+1)%ReleaseBackgroundJobMaxLockID, job.LockID,
			"Expect lock ID to be random, not auto-incremented")

		return nil
	})
	assert.NoError(t, txerr)
}

func TestCreateReleaseBackgroundJob_giveUpAfterTooManyLockIDPicks(t *testing.T) {
	ctx, err := setupCreateReleaseBackgroundJobTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		deploymentRequest2, err := CreateMockDeploymentRequestWithInProgressState(tx, ctx.org, ctx.app, nil)
		if !assert.NoError(t, err) {
			return nil
		}

		// Create a job and delete it, in order to predict what the next lock ID will be.
		job, err := CreateReleaseBackgroundJob(tx, ctx.org, ctx.app.ID, ctx.deploymentRequest)
		if !assert.NoError(t, err) {
			return nil
		}
		nextLockID := (job.LockID + 1) % ReleaseBackgroundJobMaxLockID
		err = tx.Delete(&job).Error
		if !assert.NoError(t, err) {
			return nil
		}

		// Create a job with the predicted next lock ID.
		job = ReleaseBackgroundJob{
			BaseModel: BaseModel{
				OrganizationID: ctx.org.ID,
			},
			ApplicationID:       ctx.app.ID,
			DeploymentRequestID: deploymentRequest2.ID,
			LockID:              nextLockID,
		}
		err = tx.Create(&job).Error
		if !assert.NoError(t, err) {
			return nil
		}

		// Create another job, whose autoincremented lock ID should conflict.
		_, _, err = createReleaseBackgroundJobWithDebug(tx, ctx.org, ctx.app.ID, ctx.deploymentRequest, 1)
		assert.Error(t, err, "Unable to find a free lock ID after 1 tries")

		return nil
	})
	assert.NoError(t, txerr)
}

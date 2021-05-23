package dbmodels

import (
	"testing"

	"github.com/fullstaq-labs/sqedule/server/dbutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type LoadReviewablesLatestVersionsTestContext struct {
	db       *gorm.DB
	org      Organization
	app      Application
	ruleset1 ApprovalRuleset
	ruleset2 ApprovalRuleset
}

func setupLoadReviewablesLatestVersionsTest() (LoadReviewablesLatestVersionsTestContext, error) {
	var ctx LoadReviewablesLatestVersionsTestContext
	var err error

	ctx.db, err = dbutils.SetupTestDatabase()
	if err != nil {
		return LoadReviewablesLatestVersionsTestContext{}, err
	}

	err = ctx.db.Transaction(func(tx *gorm.DB) error {
		ctx.org, err = CreateMockOrganization(tx, nil)
		if err != nil {
			return err
		}
		ctx.app, err = CreateMockApplicationWith1Version(tx, ctx.org, nil, nil)
		if err != nil {
			return err
		}
		ctx.ruleset1, err = CreateMockRulesetWith1Version(tx, ctx.org, "ruleset1", nil)
		if err != nil {
			return err
		}
		ctx.ruleset2, err = CreateMockRulesetWith1Version(tx, ctx.org, "ruleset2", nil)
		if err != nil {
			return err
		}

		return nil
	})

	return ctx, err
}

func TestLoadReviewablesLatestVersions(t *testing.T) {
	ctx, err := setupLoadReviewablesLatestVersionsTest()
	if !assert.NoError(t, err) {
		return
	}

	var versionNumber2 uint32 = 2
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		// Create binding 1
		binding1, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.org, ctx.app,
			ctx.ruleset1, nil)
		if err != nil {
			return err
		}

		// Binding 1: create version 2.1 and 2.2
		binding1Version2, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.org, ctx.app, binding1, &versionNumber2)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding1Version2, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding1Version2, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
			adjustment.AdjustmentNumber = 2
		})
		if err != nil {
			return err
		}

		// Binding 1: Create next version (no version number) and its adjustment
		binding1VersionNext, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.org, ctx.app, binding1, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding1VersionNext, nil)
		if err != nil {
			return err
		}

		// Create binding 2
		binding2, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.org, ctx.app,
			ctx.ruleset2, nil)
		if err != nil {
			return err
		}

		// Binding 2: create version 2.1, 2.2 and 2.3
		binding2Version2, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.org, ctx.app, binding2, &versionNumber2)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding2Version2, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding2Version2, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
			adjustment.AdjustmentNumber = 2
		})
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding2Version2, func(adjustment *ApplicationApprovalRulesetBindingAdjustment) {
			adjustment.AdjustmentNumber = 3
		})
		if err != nil {
			return err
		}

		// Binding 2: Create next version (no version number) and its adjustment
		binding2VersionNext, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.org, ctx.app, binding2, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingAdjustment(tx, ctx.org, binding2VersionNext, nil)
		if err != nil {
			return err
		}

		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding1, &binding2})
		if !assert.NoError(t, err) {
			return nil
		}

		// Run test: binding1's latest version should be 2.2
		assert.NotNil(t, binding1.LatestVersion)
		assert.NotNil(t, binding1.LatestVersion.VersionNumber)
		assert.NotNil(t, binding1.LatestAdjustment)
		assert.Equal(t, uint32(2), *binding1.LatestVersion.VersionNumber)
		assert.Equal(t, uint32(2), binding1.LatestAdjustment.AdjustmentNumber)

		// Run test: binding1's latest version should be 2.2
		assert.NotNil(t, binding2.LatestVersion)
		assert.NotNil(t, binding2.LatestVersion.VersionNumber)
		assert.NotNil(t, binding2.LatestAdjustment)
		assert.Equal(t, uint32(2), *binding2.LatestVersion.VersionNumber)
		assert.Equal(t, uint32(3), binding2.LatestAdjustment.AdjustmentNumber)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestLoadReviewablesLatestVersions_noVersions(t *testing.T) {
	ctx, err := setupLoadReviewablesLatestVersionsTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		binding := ApplicationApprovalRulesetBinding{
			BaseModel: BaseModel{
				OrganizationID: ctx.org.ID,
			},
			ApplicationApprovalRulesetBindingPrimaryKey: ApplicationApprovalRulesetBindingPrimaryKey{
				ApplicationID:     ctx.app.ID,
				ApprovalRulesetID: ctx.ruleset1.ID,
			},
		}
		savetx := tx.Create(&binding)
		if savetx.Error != nil {
			return savetx.Error
		}

		// Run test: latest version should not exist
		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding})
		assert.NoError(t, err)
		assert.Nil(t, binding.LatestVersion)
		assert.Nil(t, binding.LatestAdjustment)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestLoadReviewablesLatestVersions_onlyVersionIsUnfinalized(t *testing.T) {
	ctx, err := setupLoadReviewablesLatestVersionsTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		binding := ApplicationApprovalRulesetBinding{
			BaseModel: BaseModel{
				OrganizationID: ctx.org.ID,
			},
			ApplicationApprovalRulesetBindingPrimaryKey: ApplicationApprovalRulesetBindingPrimaryKey{
				ApplicationID:     ctx.app.ID,
				ApprovalRulesetID: ctx.ruleset1.ID,
			},
		}
		savetx := tx.Create(&binding)
		if savetx.Error != nil {
			return savetx.Error
		}

		_, err := CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.org, ctx.app, binding, nil)
		if err != nil {
			return err
		}

		// Run test: latest version should not exist
		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding})
		assert.NoError(t, err)
		assert.Nil(t, binding.LatestVersion)
		assert.Nil(t, binding.LatestAdjustment)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestLoadReviewablesLatestVersions_noAdjustments(t *testing.T) {
	ctx, err := setupLoadReviewablesLatestVersionsTest()
	if !assert.NoError(t, err) {
		return
	}
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		binding, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.org, ctx.app,
			ctx.ruleset1, nil)
		if err != nil {
			return err
		}
		binding.LatestVersion = nil
		binding.LatestAdjustment = nil

		// Create version 2 with no adjustments
		var versionNumber2 uint32 = 2
		_, err = CreateMockApplicationApprovalRulesetBindingVersion(tx, ctx.org, ctx.app, binding, &versionNumber2)
		if err != nil {
			return err
		}

		// Run test: latest version should be 2, adjustment version nil
		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding})
		assert.NoError(t, err)
		assert.NotNil(t, binding.LatestVersion)
		assert.NotNil(t, binding.LatestVersion.VersionNumber)
		assert.Nil(t, binding.LatestAdjustment)
		assert.Equal(t, uint32(2), *binding.LatestVersion.VersionNumber)

		return nil
	})
	assert.NoError(t, txerr)
}

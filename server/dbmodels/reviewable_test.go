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
		ctx.org, err = CreateMockOrganization(tx)
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

	var majorVersionNumber2 uint32 = 2
	txerr := ctx.db.Transaction(func(tx *gorm.DB) error {
		// Create binding 1
		binding1, err := CreateMockApplicationRulesetBindingWithEnforcingMode1Version(tx, ctx.org, ctx.app,
			ctx.ruleset1, nil)
		if err != nil {
			return err
		}

		// Binding 1: create version 2.1 and 2.2
		binding1MajorVersion2, err := CreateMockApplicationApprovalRulesetBindingMajorVersion(tx, ctx.org, ctx.app, binding1, &majorVersionNumber2)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding1MajorVersion2, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding1MajorVersion2, func(minorVersion *ApplicationApprovalRulesetBindingMinorVersion) {
			minorVersion.VersionNumber = 2
		})
		if err != nil {
			return err
		}

		// Binding 1: Create next major version (no version number) and its minor version
		binding1MajorVersionNext, err := CreateMockApplicationApprovalRulesetBindingMajorVersion(tx, ctx.org, ctx.app, binding1, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding1MajorVersionNext, nil)
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
		binding2MajorVersion2, err := CreateMockApplicationApprovalRulesetBindingMajorVersion(tx, ctx.org, ctx.app, binding2, &majorVersionNumber2)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding2MajorVersion2, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding2MajorVersion2, func(minorVersion *ApplicationApprovalRulesetBindingMinorVersion) {
			minorVersion.VersionNumber = 2
		})
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding2MajorVersion2, func(minorVersion *ApplicationApprovalRulesetBindingMinorVersion) {
			minorVersion.VersionNumber = 3
		})
		if err != nil {
			return err
		}

		// Binding 2: Create next major version (no version number) and its minor version
		binding2MajorVersionNext, err := CreateMockApplicationApprovalRulesetBindingMajorVersion(tx, ctx.org, ctx.app, binding2, nil)
		if err != nil {
			return err
		}
		_, err = CreateMockApplicationApprovalRulesetBindingMinorVersion(tx, ctx.org, binding2MajorVersionNext, nil)
		if err != nil {
			return err
		}

		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding1, &binding2})
		if !assert.NoError(t, err) {
			return nil
		}

		// Run test: binding1's latest version should be 2.2
		assert.NotNil(t, binding1.LatestMajorVersion)
		assert.NotNil(t, binding1.LatestMajorVersion.VersionNumber)
		assert.NotNil(t, binding1.LatestMinorVersion)
		assert.Equal(t, uint32(2), *binding1.LatestMajorVersion.VersionNumber)
		assert.Equal(t, uint32(2), binding1.LatestMinorVersion.VersionNumber)

		// Run test: binding1's latest version should be 2.2
		assert.NotNil(t, binding2.LatestMajorVersion)
		assert.NotNil(t, binding2.LatestMajorVersion.VersionNumber)
		assert.NotNil(t, binding2.LatestMinorVersion)
		assert.Equal(t, uint32(2), *binding2.LatestMajorVersion.VersionNumber)
		assert.Equal(t, uint32(3), binding2.LatestMinorVersion.VersionNumber)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestLoadReviewablesLatestVersions_noMajorVersions(t *testing.T) {
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
		assert.Nil(t, binding.LatestMajorVersion)
		assert.Nil(t, binding.LatestMinorVersion)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestLoadReviewablesLatestVersions_onlyMajorVersionIsUnfinalized(t *testing.T) {
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

		_, err := CreateMockApplicationApprovalRulesetBindingMajorVersion(tx, ctx.org, ctx.app, binding, nil)
		if err != nil {
			return err
		}

		// Run test: latest version should not exist
		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding})
		assert.NoError(t, err)
		assert.Nil(t, binding.LatestMajorVersion)
		assert.Nil(t, binding.LatestMinorVersion)

		return nil
	})
	assert.NoError(t, txerr)
}

func TestLoadReviewablesLatestVersions_noMinorVersions(t *testing.T) {
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
		binding.LatestMajorVersion = nil
		binding.LatestMinorVersion = nil

		// Create major version 2 with no minor versions
		var majorVersionNumber2 uint32 = 2
		_, err = CreateMockApplicationApprovalRulesetBindingMajorVersion(tx, ctx.org, ctx.app, binding, &majorVersionNumber2)
		if err != nil {
			return err
		}

		// Run test: latest major version should be 2, minor version nil
		err = LoadApplicationApprovalRulesetBindingsLatestVersions(tx, ctx.org.ID, []*ApplicationApprovalRulesetBinding{&binding})
		assert.NoError(t, err)
		assert.NotNil(t, binding.LatestMajorVersion)
		assert.NotNil(t, binding.LatestMajorVersion.VersionNumber)
		assert.Nil(t, binding.LatestMinorVersion)
		assert.Equal(t, uint32(2), *binding.LatestMajorVersion.VersionNumber)

		return nil
	})
	assert.NoError(t, txerr)
}

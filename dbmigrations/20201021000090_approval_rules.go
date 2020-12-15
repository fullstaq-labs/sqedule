package dbmigrations

import (
	"database/sql"
	"time"

	"github.com/fullstaq-labs/sqedule/dbmodels/approvalpolicy"
	"github.com/fullstaq-labs/sqedule/dbmodels/retrypolicy"
	"github.com/fullstaq-labs/sqedule/dbutils/gormigrate"
	"gorm.io/gorm"
)

func init() {
	registerDbMigration(&migration20201021000090)
}

var migration20201021000090 = gormigrate.Migration{
	ID: "20201021000090 Approval rules",
	Migrate: func(tx *gorm.DB) error {
		type Organization struct {
			ID string `gorm:"type:citext; primaryKey; not null"`
		}

		type BaseModel struct {
			OrganizationID string       `gorm:"type:citext; primaryKey; not null"`
			Organization   Organization `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
		}

		type ApprovalRulesetMinorVersion struct {
			BaseModel
			ApprovalRulesetMajorVersionID uint64 `gorm:"primaryKey; not null"`
			VersionNumber                 uint32 `gorm:"primaryKey; not null"`
		}

		// ApprovalRule ...
		type ApprovalRule struct {
			BaseModel
			ID                                uint64                      `gorm:"primaryKey; autoIncrement; not null"`
			ApprovalRulesetMajorVersionID     uint64                      `gorm:"not null"`
			ApprovalRulesetMinorVersionNumber uint32                      `gorm:"not null"`
			ApprovalRulesetMinorVersion       ApprovalRulesetMinorVersion `gorm:"foreignKey:OrganizationID,ApprovalRulesetMajorVersionID,ApprovalRulesetMinorVersionNumber; references:OrganizationID,ApprovalRulesetMajorVersionID,VersionNumber; constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
			Enabled                           bool                        `gorm:"not null; default:true"`
			CreatedAt                         time.Time                   `gorm:"not null"`
		}

		// HTTPApiApprovalRule ...
		type HTTPApiApprovalRule struct {
			ApprovalRule
			URL              string `gorm:"not null"`
			Username         sql.NullString
			Password         sql.NullString
			TLSCaCertificate sql.NullString
			RetryPolicy      retrypolicy.Policy `gorm:"type:retry_policy; not null"`
			RetryLimit       int                `gorm:"not null; default:1; check:((retry_policy = 'retry_on_fail') = (retry_limit IS NOT NULL))"`
		}

		// ScheduleApprovalRule ...
		type ScheduleApprovalRule struct {
			ApprovalRule
			BeginTime    sql.NullString `gorm:"check:((begin_time IS NULL) = (end_time IS NULL))"`
			EndTime      sql.NullString
			DaysOfWeek   sql.NullString
			DaysOfMonth  sql.NullString
			MonthsOfYear sql.NullString
		}

		// ManualApprovalRule ...
		type ManualApprovalRule struct {
			ApprovalRule
			ApprovalPolicy approvalpolicy.Policy `gorm:"type:approval_policy; not null"`
			Minimum        sql.NullInt32         `gorm:"check:((approval_policy = 'minimum') = (minimum IS NOT NULL))"`
		}

		err := tx.Exec("CREATE TYPE retry_policy AS ENUM " +
			"('never', 'retry_on_fail')").Error
		if err != nil {
			return err
		}

		err = tx.Exec("CREATE TYPE approval_policy AS ENUM " +
			"('any', 'all', 'minimum')").Error
		if err != nil {
			return err
		}

		return tx.AutoMigrate(&HTTPApiApprovalRule{}, &ScheduleApprovalRule{},
			&ManualApprovalRule{})
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Migrator().DropTable("http_api_approval_rules",
			"schedule_approval_rules", "manual_approval_rules")
		if err != nil {
			return err
		}

		err = tx.Exec("DROP TYPE approval_policy").Error
		if err != nil {
			return err
		}

		return tx.Exec("DROP TYPE retry_policy").Error
	},
}
